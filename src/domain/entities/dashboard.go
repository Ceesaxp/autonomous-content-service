package entities

import (
	"time"

	"github.com/google/uuid"
)

// DashboardView represents different views available in the dashboard
type DashboardView string

const (
	DashboardViewOverview   DashboardView = "Overview"
	DashboardViewProjects   DashboardView = "Projects" 
	DashboardViewContent    DashboardView = "Content"
	DashboardViewBilling    DashboardView = "Billing"
	DashboardViewAnalytics  DashboardView = "Analytics"
	DashboardViewMessages   DashboardView = "Messages"
	DashboardViewSettings   DashboardView = "Settings"
)

// NotificationType represents different types of notifications
type NotificationType string

const (
	NotificationTypeContentReady    NotificationType = "ContentReady"
	NotificationTypeRevisionRequest NotificationType = "RevisionRequest"
	NotificationTypePaymentDue      NotificationType = "PaymentDue"
	NotificationTypeProjectUpdate   NotificationType = "ProjectUpdate"
	NotificationTypeMessage         NotificationType = "Message"
	NotificationTypeDeadlineAlert   NotificationType = "DeadlineAlert"
)

// NotificationPriority represents the priority level of notifications
type NotificationPriority string

const (
	NotificationPriorityHigh   NotificationPriority = "High"
	NotificationPriorityMedium NotificationPriority = "Medium"
	NotificationPriorityLow    NotificationPriority = "Low"
)

// DashboardNotification represents a notification in the dashboard
type DashboardNotification struct {
	NotificationID uuid.UUID            `json:"notificationId"`
	ClientID       uuid.UUID            `json:"clientId"`
	ProjectID      *uuid.UUID           `json:"projectId,omitempty"`
	Type           NotificationType     `json:"type"`
	Priority       NotificationPriority `json:"priority"`
	Title          string               `json:"title"`
	Message        string               `json:"message"`
	ActionURL      string               `json:"actionUrl,omitempty"`
	IsRead         bool                 `json:"isRead"`
	CreatedAt      time.Time            `json:"createdAt"`
	ReadAt         *time.Time           `json:"readAt,omitempty"`
}

// NewDashboardNotification creates a new dashboard notification
func NewDashboardNotification(clientID uuid.UUID, notificationType NotificationType, title, message string) *DashboardNotification {
	return &DashboardNotification{
		NotificationID: uuid.New(),
		ClientID:       clientID,
		Type:           notificationType,
		Priority:       NotificationPriorityMedium,
		Title:          title,
		Message:        message,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}
}

// MarkAsRead marks the notification as read
func (n *DashboardNotification) MarkAsRead() {
	if !n.IsRead {
		n.IsRead = true
		now := time.Now()
		n.ReadAt = &now
	}
}

// ProjectMilestone represents a milestone in a project
type ProjectMilestone struct {
	MilestoneID uuid.UUID `json:"milestoneId"`
	ProjectID   uuid.UUID `json:"projectId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"dueDate"`
	IsCompleted bool      `json:"isCompleted"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// NewProjectMilestone creates a new project milestone
func NewProjectMilestone(projectID uuid.UUID, title, description string, dueDate time.Time) *ProjectMilestone {
	return &ProjectMilestone{
		MilestoneID: uuid.New(),
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		IsCompleted: false,
		CreatedAt:   time.Now(),
	}
}

// Complete marks the milestone as completed
func (m *ProjectMilestone) Complete() {
	if !m.IsCompleted {
		m.IsCompleted = true
		now := time.Now()
		m.CompletedAt = &now
	}
}

// ContentApprovalStatus represents the approval status of content
type ContentApprovalStatus string

const (
	ContentApprovalPending  ContentApprovalStatus = "Pending"
	ContentApprovalApproved ContentApprovalStatus = "Approved"
	ContentApprovalRejected ContentApprovalStatus = "Rejected"
	ContentApprovalRevision ContentApprovalStatus = "Revision"
)

// ContentApproval represents an approval request for content
type ContentApproval struct {
	ApprovalID uuid.UUID             `json:"approvalId"`
	ContentID  uuid.UUID             `json:"contentId"`
	ProjectID  uuid.UUID             `json:"projectId"`
	ClientID   uuid.UUID             `json:"clientId"`
	Status     ContentApprovalStatus `json:"status"`
	Feedback   string                `json:"feedback,omitempty"`
	CreatedAt  time.Time             `json:"createdAt"`
	UpdatedAt  time.Time             `json:"updatedAt"`
	ApprovedAt *time.Time            `json:"approvedAt,omitempty"`
}

// NewContentApproval creates a new content approval request
func NewContentApproval(contentID, projectID, clientID uuid.UUID) *ContentApproval {
	return &ContentApproval{
		ApprovalID: uuid.New(),
		ContentID:  contentID,
		ProjectID:  projectID,
		ClientID:   clientID,
		Status:     ContentApprovalPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Approve approves the content
func (a *ContentApproval) Approve(feedback string) {
	a.Status = ContentApprovalApproved
	a.Feedback = feedback
	now := time.Now()
	a.ApprovedAt = &now
	a.UpdatedAt = now
}

// Reject rejects the content with feedback
func (a *ContentApproval) Reject(feedback string) {
	a.Status = ContentApprovalRejected
	a.Feedback = feedback
	a.UpdatedAt = time.Now()
}

// RequestRevision requests revision with feedback
func (a *ContentApproval) RequestRevision(feedback string) {
	a.Status = ContentApprovalRevision
	a.Feedback = feedback
	a.UpdatedAt = time.Now()
}

// RevisionRequest represents a request for content revision
type RevisionRequest struct {
	RequestID   uuid.UUID `json:"requestId"`
	ContentID   uuid.UUID `json:"contentId"`
	ProjectID   uuid.UUID `json:"projectId"`
	ClientID    uuid.UUID `json:"clientId"`
	Reason      string    `json:"reason"`
	Details     string    `json:"details"`
	IsCompleted bool      `json:"isCompleted"`
	CreatedAt   time.Time `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

// NewRevisionRequest creates a new revision request
func NewRevisionRequest(contentID, projectID, clientID uuid.UUID, reason, details string) *RevisionRequest {
	return &RevisionRequest{
		RequestID:   uuid.New(),
		ContentID:   contentID,
		ProjectID:   projectID,
		ClientID:    clientID,
		Reason:      reason,
		Details:     details,
		IsCompleted: false,
		CreatedAt:   time.Now(),
	}
}

// Complete marks the revision request as completed
func (r *RevisionRequest) Complete() {
	if !r.IsCompleted {
		r.IsCompleted = true
		now := time.Now()
		r.CompletedAt = &now
	}
}

// MessageThread represents a communication thread between client and system
type MessageThread struct {
	ThreadID    uuid.UUID `json:"threadId"`
	ProjectID   uuid.UUID `json:"projectId"`
	ClientID    uuid.UUID `json:"clientId"`
	Subject     string    `json:"subject"`
	IsActive    bool      `json:"isActive"`
	LastMessage *time.Time `json:"lastMessage,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// NewMessageThread creates a new message thread
func NewMessageThread(projectID, clientID uuid.UUID, subject string) *MessageThread {
	return &MessageThread{
		ThreadID:  uuid.New(),
		ProjectID: projectID,
		ClientID:  clientID,
		Subject:   subject,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
}

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeClient MessageType = "Client"
	MessageTypeSystem MessageType = "System"
	MessageTypeAuto   MessageType = "Auto"
)

// DashboardMessage represents a message in a thread
type DashboardMessage struct {
	MessageID uuid.UUID   `json:"messageId"`
	ThreadID  uuid.UUID   `json:"threadId"`
	Type      MessageType `json:"type"`
	Content   string      `json:"content"`
	IsRead    bool        `json:"isRead"`
	CreatedAt time.Time   `json:"createdAt"`
	ReadAt    *time.Time  `json:"readAt,omitempty"`
}

// NewDashboardMessage creates a new dashboard message
func NewDashboardMessage(threadID uuid.UUID, messageType MessageType, content string) *DashboardMessage {
	return &DashboardMessage{
		MessageID: uuid.New(),
		ThreadID:  threadID,
		Type:      messageType,
		Content:   content,
		IsRead:    false,
		CreatedAt: time.Now(),
	}
}

// MarkAsRead marks the message as read
func (m *DashboardMessage) MarkAsRead() {
	if !m.IsRead {
		m.IsRead = true
		now := time.Now()
		m.ReadAt = &now
	}
}

// ProjectAnalytics represents analytics data for a project
type ProjectAnalytics struct {
	ProjectID        uuid.UUID              `json:"projectId"`
	TotalTasks       int                    `json:"totalTasks"`
	CompletedTasks   int                    `json:"completedTasks"`
	PendingTasks     int                    `json:"pendingTasks"`
	ProgressPercent  float64                `json:"progressPercent"`
	TimeSpent        time.Duration          `json:"timeSpent"`
	EstimatedTime    time.Duration          `json:"estimatedTime"`
	DaysToDeadline   int                    `json:"daysToDeadline"`
	ContentDelivered int                    `json:"contentDelivered"`
	RevisionRequests int                    `json:"revisionRequests"`
	ClientSatisfaction float64              `json:"clientSatisfaction"`
	CustomMetrics    map[string]interface{} `json:"customMetrics,omitempty"`
	LastUpdated      time.Time              `json:"lastUpdated"`
}

// NewProjectAnalytics creates new project analytics
func NewProjectAnalytics(projectID uuid.UUID) *ProjectAnalytics {
	return &ProjectAnalytics{
		ProjectID:      projectID,
		CustomMetrics:  make(map[string]interface{}),
		LastUpdated:    time.Now(),
	}
}

// UpdateProgress updates the progress metrics
func (pa *ProjectAnalytics) UpdateProgress(completed, total int) {
	pa.CompletedTasks = completed
	pa.TotalTasks = total
	pa.PendingTasks = total - completed
	if total > 0 {
		pa.ProgressPercent = float64(completed) / float64(total) * 100
	}
	pa.LastUpdated = time.Now()
}

// BillingHistory represents billing and payment history
type BillingHistory struct {
	BillingID      uuid.UUID              `json:"billingId"`
	ProjectID      uuid.UUID              `json:"projectId"`
	ClientID       uuid.UUID              `json:"clientId"`
	InvoiceNumber  string                 `json:"invoiceNumber"`
	Amount         Money                  `json:"amount"`
	Status         PaymentStatus          `json:"status"`
	DueDate        time.Time              `json:"dueDate"`
	PaidDate       *time.Time             `json:"paidDate,omitempty"`
	Description    string                 `json:"description"`
	LineItems      []BillingLineItem      `json:"lineItems,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
}

// BillingLineItem represents an individual item in a billing record
type BillingLineItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   Money   `json:"unitPrice"`
	Total       Money   `json:"total"`
}

// NewBillingHistory creates a new billing history record
func NewBillingHistory(projectID, clientID uuid.UUID, invoiceNumber string, amount Money, dueDate time.Time) *BillingHistory {
	return &BillingHistory{
		BillingID:     uuid.New(),
		ProjectID:     projectID,
		ClientID:      clientID,
		InvoiceNumber: invoiceNumber,
		Amount:        amount,
		Status:        PaymentStatusPending,
		DueDate:       dueDate,
		Metadata:      make(map[string]interface{}),
		CreatedAt:     time.Now(),
	}
}

// MarkAsPaid marks the billing record as paid
func (bh *BillingHistory) MarkAsPaid() {
	bh.Status = PaymentStatusCompleted
	now := time.Now()
	bh.PaidDate = &now
}