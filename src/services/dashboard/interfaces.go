package dashboard

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// DashboardService defines the interface for dashboard business logic
type DashboardService interface {
	// Dashboard Overview
	GetDashboardSummary(ctx context.Context, clientID uuid.UUID) (*repositories.DashboardSummary, error)
	RefreshDashboardData(ctx context.Context, clientID uuid.UUID) error

	// Project Management
	GetProjectsOverview(ctx context.Context, clientID uuid.UUID) ([]*ProjectOverview, error)
	GetProjectDetails(ctx context.Context, projectID uuid.UUID) (*ProjectDetails, error)
	UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status entities.ProjectStatus) error
	
	// Content Management
	GetContentApprovals(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.ContentApproval, error)
	ApproveContent(ctx context.Context, approvalID uuid.UUID, feedback string) error
	RejectContent(ctx context.Context, approvalID uuid.UUID, feedback string) error
	RequestContentRevision(ctx context.Context, approvalID uuid.UUID, feedback string) error
	
	// Communication
	GetMessageThreads(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.MessageThread, error)
	CreateMessageThread(ctx context.Context, projectID, clientID uuid.UUID, subject string) (*entities.MessageThread, error)
	SendMessage(ctx context.Context, threadID uuid.UUID, messageType entities.MessageType, content string) (*entities.DashboardMessage, error)
	GetThreadMessages(ctx context.Context, threadID uuid.UUID, limit, offset int) ([]*entities.DashboardMessage, error)
	MarkMessagesAsRead(ctx context.Context, threadID uuid.UUID) error
	
	// Notifications
	GetNotifications(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.DashboardNotification, error)
	CreateNotification(ctx context.Context, clientID uuid.UUID, notificationType entities.NotificationType, title, message string) error
	MarkNotificationAsRead(ctx context.Context, notificationID uuid.UUID) error
	GetUnreadNotificationCount(ctx context.Context, clientID uuid.UUID) (int, error)
	
	// Analytics
	GetProjectAnalytics(ctx context.Context, projectID uuid.UUID) (*entities.ProjectAnalytics, error)
	GetClientAnalytics(ctx context.Context, clientID uuid.UUID, fromDate, toDate time.Time) (*ClientAnalytics, error)
	GeneratePerformanceReport(ctx context.Context, clientID uuid.UUID, reportType ReportType) (*PerformanceReport, error)
	
	// Billing
	GetBillingHistory(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.BillingHistory, error)
	GetOutstandingInvoices(ctx context.Context, clientID uuid.UUID) ([]*entities.BillingHistory, error)
	CreateInvoice(ctx context.Context, projectID, clientID uuid.UUID, amount entities.Money, description string) (*entities.BillingHistory, error)
	MarkInvoiceAsPaid(ctx context.Context, billingID uuid.UUID) error
}

// ProjectOverview represents a summary view of a project for dashboard
type ProjectOverview struct {
	ProjectID       uuid.UUID                `json:"projectId"`
	Title           string                   `json:"title"`
	Status          entities.ProjectStatus   `json:"status"`
	Priority        entities.Priority        `json:"priority"`
	Progress        float64                  `json:"progress"`
	Deadline        time.Time                `json:"deadline"`
	DaysRemaining   int                      `json:"daysRemaining"`
	Budget          entities.Money           `json:"budget"`
	ContentCount    int                      `json:"contentCount"`
	PendingApprovals int                     `json:"pendingApprovals"`
	LastUpdate      time.Time                `json:"lastUpdate"`
}

// ProjectDetails represents detailed project information
type ProjectDetails struct {
	Project          *entities.Project            `json:"project"`
	Milestones       []*entities.ProjectMilestone `json:"milestones"`
	ContentApprovals []*entities.ContentApproval  `json:"contentApprovals"`
	Analytics        *entities.ProjectAnalytics   `json:"analytics"`
	MessageThreads   []*entities.MessageThread    `json:"messageThreads"`
	BillingHistory   []*entities.BillingHistory   `json:"billingHistory"`
}

// ClientAnalytics represents comprehensive analytics for a client
type ClientAnalytics struct {
	ClientID           uuid.UUID                        `json:"clientId"`
	TotalProjects      int                              `json:"totalProjects"`
	ActiveProjects     int                              `json:"activeProjects"`
	CompletedProjects  int                              `json:"completedProjects"`
	TotalContentPieces int                              `json:"totalContentPieces"`
	AverageProjectTime time.Duration                    `json:"averageProjectTime"`
	TotalSpent         entities.Money                   `json:"totalSpent"`
	ProjectsByStatus   map[entities.ProjectStatus]int   `json:"projectsByStatus"`
	ContentByType      map[entities.ContentType]int     `json:"contentByType"`
	MonthlyActivity    []MonthlyActivity                `json:"monthlyActivity"`
	PerformanceMetrics PerformanceMetrics               `json:"performanceMetrics"`
	FromDate           time.Time                        `json:"fromDate"`
	ToDate             time.Time                        `json:"toDate"`
}

// MonthlyActivity represents activity data for a specific month
type MonthlyActivity struct {
	Month           string         `json:"month"`
	ProjectsStarted int            `json:"projectsStarted"`
	ProjectsCompleted int          `json:"projectsCompleted"`
	ContentDelivered int           `json:"contentDelivered"`
	AmountSpent     entities.Money `json:"amountSpent"`
}

// PerformanceMetrics represents key performance indicators
type PerformanceMetrics struct {
	OnTimeDeliveryRate    float64 `json:"onTimeDeliveryRate"`
	ClientSatisfactionAvg float64 `json:"clientSatisfactionAvg"`
	RevisionRequestRate   float64 `json:"revisionRequestRate"`
	AverageResponseTime   string  `json:"averageResponseTime"`
}

// ReportType represents different types of performance reports
type ReportType string

const (
	ReportTypeProjectSummary   ReportType = "ProjectSummary"
	ReportTypeContentDelivery  ReportType = "ContentDelivery"
	ReportTypeFinancialSummary ReportType = "FinancialSummary"
	ReportTypePerformanceKPI   ReportType = "PerformanceKPI"
)

// PerformanceReport represents a generated performance report
type PerformanceReport struct {
	ReportID    uuid.UUID              `json:"reportId"`
	ClientID    uuid.UUID              `json:"clientId"`
	Type        ReportType             `json:"type"`
	Title       string                 `json:"title"`
	Summary     string                 `json:"summary"`
	Data        map[string]interface{} `json:"data"`
	Charts      []ChartData            `json:"charts"`
	Insights    []string               `json:"insights"`
	Recommendations []string           `json:"recommendations"`
	GeneratedAt time.Time              `json:"generatedAt"`
}

// ChartData represents data for dashboard charts
type ChartData struct {
	Type   string                 `json:"type"`  // "line", "bar", "pie", "doughnut"
	Title  string                 `json:"title"`
	Labels []string               `json:"labels"`
	Data   []interface{}          `json:"data"`
	Options map[string]interface{} `json:"options,omitempty"`
}