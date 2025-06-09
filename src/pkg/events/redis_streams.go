package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// EventPublisher interface for publishing events
type EventPublisher interface {
	Publish(ctx context.Context, stream string, event Event) error
	Close() error
}

// EventConsumer interface for consuming events
type EventConsumer interface {
	Subscribe(ctx context.Context, stream, consumerGroup string, handler EventHandler) error
	Close() error
}

// EventHandler function type for handling events
type EventHandler func(ctx context.Context, event Event) error

// Event represents a system event
type Event struct {
	ID        string                 `json:"event_id"`
	Type      string                 `json:"event_type"`
	Source    string                 `json:"source_service"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]string      `json:"metadata"`
}

// RedisStreamsClient implements event publishing and consumption using Redis Streams
type RedisStreamsClient struct {
	client       *redis.Client
	consumerName string
}

// NewRedisStreamsClient creates a new Redis Streams event client
func NewRedisStreamsClient(redisAddr, password string, db int) *RedisStreamsClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: password,
		DB:       db,
	})

	// Generate unique consumer name
	consumerName := fmt.Sprintf("consumer-%s", uuid.New().String()[:8])

	return &RedisStreamsClient{
		client:       rdb,
		consumerName: consumerName,
	}
}

// Publish publishes an event to a Redis stream
func (r *RedisStreamsClient) Publish(ctx context.Context, stream string, event Event) error {
	// Ensure event has required fields
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Convert event to map for Redis
	eventData := map[string]interface{}{
		"event_id":     event.ID,
		"event_type":   event.Type,
		"source":       event.Source,
		"timestamp":    event.Timestamp.Format(time.RFC3339),
		"metadata":     marshalMap(event.Metadata),
	}

	// Add payload fields
	if event.Payload != nil {
		payloadJSON, err := json.Marshal(event.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal event payload: %w", err)
		}
		eventData["payload"] = string(payloadJSON)
	}

	// Publish to Redis stream
	result := r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: eventData,
	})

	if err := result.Err(); err != nil {
		return fmt.Errorf("failed to publish event to stream %s: %w", stream, err)
	}

	log.Printf("Published event %s to stream %s (message ID: %s)", event.ID, stream, result.Val())
	return nil
}

// Subscribe subscribes to a Redis stream and processes events with the given handler
func (r *RedisStreamsClient) Subscribe(ctx context.Context, stream, consumerGroup string, handler EventHandler) error {
	// Create consumer group if it doesn't exist
	_, err := r.client.XGroupCreateMkStream(ctx, stream, consumerGroup, "0").Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	log.Printf("Subscribing to stream %s with consumer group %s", stream, consumerGroup)

	// Start consuming messages
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Read messages from the stream
			results, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    consumerGroup,
				Consumer: r.consumerName,
				Streams:  []string{stream, ">"},
				Count:    10,
				Block:    time.Second * 5,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// No new messages, continue
					continue
				}
				log.Printf("Error reading from stream: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}

			// Process each message
			for _, result := range results {
				for _, message := range result.Messages {
					event, err := r.parseEvent(message.Values)
					if err != nil {
						log.Printf("Error parsing event: %v", err)
						continue
					}

					// Handle the event
					if err := handler(ctx, event); err != nil {
						log.Printf("Error handling event %s: %v", event.ID, err)
						continue
					}

					// Acknowledge the message
					r.client.XAck(ctx, stream, consumerGroup, message.ID)
				}
			}
		}
	}
}

// Close closes the Redis client connection
func (r *RedisStreamsClient) Close() error {
	return r.client.Close()
}

// parseEvent converts Redis message values to Event struct
func (r *RedisStreamsClient) parseEvent(values map[string]interface{}) (Event, error) {
	event := Event{}

	// Extract basic fields
	if id, ok := values["event_id"].(string); ok {
		event.ID = id
	}
	if eventType, ok := values["event_type"].(string); ok {
		event.Type = eventType
	}
	if source, ok := values["source"].(string); ok {
		event.Source = source
	}
	if timestampStr, ok := values["timestamp"].(string); ok {
		if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			event.Timestamp = timestamp
		}
	}

	// Parse metadata
	if metadataStr, ok := values["metadata"].(string); ok {
		if err := json.Unmarshal([]byte(metadataStr), &event.Metadata); err != nil {
			log.Printf("Error parsing metadata: %v", err)
		}
	}

	// Parse payload
	if payloadStr, ok := values["payload"].(string); ok {
		if err := json.Unmarshal([]byte(payloadStr), &event.Payload); err != nil {
			log.Printf("Error parsing payload: %v", err)
		}
	}

	return event, nil
}

// marshalMap converts a map to JSON string
func marshalMap(m map[string]string) string {
	if m == nil {
		return "{}"
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// CreateEvent creates a new event with the given parameters
func CreateEvent(eventType, source string, payload map[string]interface{}) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
		Metadata:  make(map[string]string),
	}
}

// WithMetadata adds metadata to an event
func (e Event) WithMetadata(key, value string) Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}