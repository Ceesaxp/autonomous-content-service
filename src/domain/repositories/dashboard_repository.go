package repositories

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// DashboardRepository defines the interface for dashboard data operations
type DashboardRepository interface {
	// Notifications
	CreateNotification(ctx context.Context, notification *entities.DashboardNotification) error
	GetNotificationsByClientID(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.DashboardNotification, error)
	GetUnreadNotificationCount(ctx context.Context, clientID uuid.UUID) (int, error)
	MarkNotificationAsRead(ctx context.Context, notificationID uuid.UUID) error
	DeleteNotification(ctx context.Context, notificationID uuid.UUID) error

	// Milestones
	CreateMilestone(ctx context.Context, milestone *entities.ProjectMilestone) error
	GetMilestonesByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entities.ProjectMilestone, error)
	UpdateMilestone(ctx context.Context, milestone *entities.ProjectMilestone) error
	CompleteMilestone(ctx context.Context, milestoneID uuid.UUID) error
	DeleteMilestone(ctx context.Context, milestoneID uuid.UUID) error

	// Content Approvals
	CreateContentApproval(ctx context.Context, approval *entities.ContentApproval) error
	GetContentApprovalsByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entities.ContentApproval, error)
	GetContentApprovalsByClientID(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.ContentApproval, error)
	UpdateContentApproval(ctx context.Context, approval *entities.ContentApproval) error
	GetContentApprovalByID(ctx context.Context, approvalID uuid.UUID) (*entities.ContentApproval, error)

	// Revision Requests
	CreateRevisionRequest(ctx context.Context, request *entities.RevisionRequest) error
	GetRevisionRequestsByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entities.RevisionRequest, error)
	GetRevisionRequestsByClientID(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.RevisionRequest, error)
	UpdateRevisionRequest(ctx context.Context, request *entities.RevisionRequest) error
	CompleteRevisionRequest(ctx context.Context, requestID uuid.UUID) error

	// Message Threads
	CreateMessageThread(ctx context.Context, thread *entities.MessageThread) error
	GetMessageThreadsByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entities.MessageThread, error)
	GetMessageThreadsByClientID(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.MessageThread, error)
	UpdateMessageThread(ctx context.Context, thread *entities.MessageThread) error
	GetMessageThreadByID(ctx context.Context, threadID uuid.UUID) (*entities.MessageThread, error)

	// Messages
	CreateMessage(ctx context.Context, message *entities.DashboardMessage) error
	GetMessagesByThreadID(ctx context.Context, threadID uuid.UUID, limit, offset int) ([]*entities.DashboardMessage, error)
	MarkMessageAsRead(ctx context.Context, messageID uuid.UUID) error
	GetUnreadMessageCount(ctx context.Context, clientID uuid.UUID) (int, error)

	// Analytics
	CreateProjectAnalytics(ctx context.Context, analytics *entities.ProjectAnalytics) error
	GetProjectAnalytics(ctx context.Context, projectID uuid.UUID) (*entities.ProjectAnalytics, error)
	UpdateProjectAnalytics(ctx context.Context, analytics *entities.ProjectAnalytics) error
	GetClientAnalyticsSummary(ctx context.Context, clientID uuid.UUID, fromDate, toDate time.Time) (map[string]interface{}, error)

	// Billing History
	CreateBillingHistory(ctx context.Context, billing *entities.BillingHistory) error
	GetBillingHistoryByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entities.BillingHistory, error)
	GetBillingHistoryByClientID(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.BillingHistory, error)
	UpdateBillingHistory(ctx context.Context, billing *entities.BillingHistory) error
	GetOutstandingInvoices(ctx context.Context, clientID uuid.UUID) ([]*entities.BillingHistory, error)

	// Dashboard Summary
	GetDashboardSummary(ctx context.Context, clientID uuid.UUID) (*DashboardSummary, error)
	GetProjectStatusSummary(ctx context.Context, clientID uuid.UUID) (map[entities.ProjectStatus]int, error)
}

// DashboardSummary represents a summary of dashboard data for a client
type DashboardSummary struct {
	ClientID             uuid.UUID                            `json:"clientId"`
	ActiveProjects       int                                  `json:"activeProjects"`
	CompletedProjects    int                                  `json:"completedProjects"`
	PendingApprovals     int                                  `json:"pendingApprovals"`
	UnreadNotifications  int                                  `json:"unreadNotifications"`
	UnreadMessages       int                                  `json:"unreadMessages"`
	OutstandingBalance   entities.Money                       `json:"outstandingBalance"`
	ProjectStatusBreakdown map[entities.ProjectStatus]int     `json:"projectStatusBreakdown"`
	RecentActivity       []DashboardActivity                  `json:"recentActivity"`
	UpcomingDeadlines    []DashboardDeadline                  `json:"upcomingDeadlines"`
	LastUpdated          time.Time                            `json:"lastUpdated"`
}

// DashboardActivity represents recent activity items
type DashboardActivity struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	ProjectID   uuid.UUID `json:"projectId,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// DashboardDeadline represents upcoming deadlines
type DashboardDeadline struct {
	ProjectID   uuid.UUID `json:"projectId"`
	ProjectName string    `json:"projectName"`
	Deadline    time.Time `json:"deadline"`
	DaysLeft    int       `json:"daysLeft"`
	IsOverdue   bool      `json:"isOverdue"`
}