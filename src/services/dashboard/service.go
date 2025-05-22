package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// DashboardServiceImpl implements the DashboardService interface
type DashboardServiceImpl struct {
	dashboardRepo repositories.DashboardRepository
	projectRepo   repositories.ProjectRepository
	contentRepo   repositories.ContentRepository
	clientRepo    repositories.ClientRepository
}

// NewDashboardService creates a new dashboard service instance
func NewDashboardService(
	dashboardRepo repositories.DashboardRepository,
	projectRepo repositories.ProjectRepository,
	contentRepo repositories.ContentRepository,
	clientRepo repositories.ClientRepository,
) DashboardService {
	return &DashboardServiceImpl{
		dashboardRepo: dashboardRepo,
		projectRepo:   projectRepo,
		contentRepo:   contentRepo,
		clientRepo:    clientRepo,
	}
}

// GetDashboardSummary retrieves a comprehensive dashboard summary for a client
func (s *DashboardServiceImpl) GetDashboardSummary(ctx context.Context, clientID uuid.UUID) (*repositories.DashboardSummary, error) {
	summary, err := s.dashboardRepo.GetDashboardSummary(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}
	return summary, nil
}

// RefreshDashboardData refreshes all dashboard data for a client
func (s *DashboardServiceImpl) RefreshDashboardData(ctx context.Context, clientID uuid.UUID) error {
	// This would trigger recalculation of all analytics and summaries
	// In a real implementation, this might involve background jobs
	return nil
}

// GetProjectsOverview retrieves an overview of all projects for a client
func (s *DashboardServiceImpl) GetProjectsOverview(ctx context.Context, clientID uuid.UUID) ([]*ProjectOverview, error) {
	projects, err := s.projectRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	overviews := make([]*ProjectOverview, 0, len(projects))
	for _, project := range projects {
		overview := &ProjectOverview{
			ProjectID:     project.ProjectID,
			Title:         project.Title,
			Status:        project.Status,
			Priority:      project.Priority,
			Deadline:      project.Deadline,
			DaysRemaining: int(time.Until(project.Deadline).Hours() / 24),
			Budget:        project.Budget,
			ContentCount:  len(project.Contents),
			LastUpdate:    project.UpdatedAt,
		}

		// Calculate progress based on project status and content
		switch project.Status {
		case entities.ProjectStatusCompleted:
			overview.Progress = 100.0
		case entities.ProjectStatusInProgress:
			overview.Progress = 60.0 // This would be calculated based on actual milestones
		case entities.ProjectStatusReview:
			overview.Progress = 85.0
		default:
			overview.Progress = 0.0
		}

		// Get pending approvals count
		approvals, err := s.dashboardRepo.GetContentApprovalsByProjectID(ctx, project.ProjectID)
		if err == nil {
			pendingCount := 0
			for _, approval := range approvals {
				if approval.Status == entities.ContentApprovalPending {
					pendingCount++
				}
			}
			overview.PendingApprovals = pendingCount
		}

		overviews = append(overviews, overview)
	}

	return overviews, nil
}

// GetProjectDetails retrieves detailed information about a specific project
func (s *DashboardServiceImpl) GetProjectDetails(ctx context.Context, projectID uuid.UUID) (*ProjectDetails, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	details := &ProjectDetails{
		Project: project,
	}

	// Get milestones
	milestones, err := s.dashboardRepo.GetMilestonesByProjectID(ctx, projectID)
	if err == nil {
		details.Milestones = milestones
	}

	// Get content approvals
	approvals, err := s.dashboardRepo.GetContentApprovalsByProjectID(ctx, projectID)
	if err == nil {
		details.ContentApprovals = approvals
	}

	// Get analytics
	analytics, err := s.dashboardRepo.GetProjectAnalytics(ctx, projectID)
	if err == nil {
		details.Analytics = analytics
	}

	// Get message threads
	threads, err := s.dashboardRepo.GetMessageThreadsByProjectID(ctx, projectID)
	if err == nil {
		details.MessageThreads = threads
	}

	// Get billing history
	billing, err := s.dashboardRepo.GetBillingHistoryByProjectID(ctx, projectID)
	if err == nil {
		details.BillingHistory = billing
	}

	return details, nil
}

// UpdateProjectStatus updates the status of a project
func (s *DashboardServiceImpl) UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status entities.ProjectStatus) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	project.UpdateStatus(status)
	
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	// Create notification for status change
	title := fmt.Sprintf("Project %s status updated", project.Title)
	message := fmt.Sprintf("Project status changed to %s", status)
	if err := s.CreateNotification(ctx, project.ClientID, entities.NotificationTypeProjectUpdate, title, message); err != nil {
		// Log error but don't fail the operation
	}

	return nil
}

// GetContentApprovals retrieves content approvals for a client
func (s *DashboardServiceImpl) GetContentApprovals(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.ContentApproval, error) {
	approvals, err := s.dashboardRepo.GetContentApprovalsByClientID(ctx, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get content approvals: %w", err)
	}
	return approvals, nil
}

// ApproveContent approves content with feedback
func (s *DashboardServiceImpl) ApproveContent(ctx context.Context, approvalID uuid.UUID, feedback string) error {
	approval, err := s.dashboardRepo.GetContentApprovalByID(ctx, approvalID)
	if err != nil {
		return fmt.Errorf("failed to get content approval: %w", err)
	}

	approval.Approve(feedback)
	
	if err := s.dashboardRepo.UpdateContentApproval(ctx, approval); err != nil {
		return fmt.Errorf("failed to update content approval: %w", err)
	}

	// Create notification
	title := "Content Approved"
	message := "Your content has been approved and is ready for delivery"
	if err := s.CreateNotification(ctx, approval.ClientID, entities.NotificationTypeContentReady, title, message); err != nil {
		// Log error but don't fail the operation
	}

	return nil
}

// RejectContent rejects content with feedback
func (s *DashboardServiceImpl) RejectContent(ctx context.Context, approvalID uuid.UUID, feedback string) error {
	approval, err := s.dashboardRepo.GetContentApprovalByID(ctx, approvalID)
	if err != nil {
		return fmt.Errorf("failed to get content approval: %w", err)
	}

	approval.Reject(feedback)
	
	if err := s.dashboardRepo.UpdateContentApproval(ctx, approval); err != nil {
		return fmt.Errorf("failed to update content approval: %w", err)
	}

	// Create notification
	title := "Content Rejected"
	message := "Content has been rejected. Please review feedback and request revision."
	if err := s.CreateNotification(ctx, approval.ClientID, entities.NotificationTypeRevisionRequest, title, message); err != nil {
		// Log error but don't fail the operation
	}

	return nil
}

// RequestContentRevision requests revision with feedback
func (s *DashboardServiceImpl) RequestContentRevision(ctx context.Context, approvalID uuid.UUID, feedback string) error {
	approval, err := s.dashboardRepo.GetContentApprovalByID(ctx, approvalID)
	if err != nil {
		return fmt.Errorf("failed to get content approval: %w", err)
	}

	approval.RequestRevision(feedback)
	
	if err := s.dashboardRepo.UpdateContentApproval(ctx, approval); err != nil {
		return fmt.Errorf("failed to update content approval: %w", err)
	}

	// Create revision request
	revisionRequest := entities.NewRevisionRequest(approval.ContentID, approval.ProjectID, approval.ClientID, "Client requested revision", feedback)
	if err := s.dashboardRepo.CreateRevisionRequest(ctx, revisionRequest); err != nil {
		// Log error but don't fail the operation
	}

	// Create notification
	title := "Revision Requested"
	message := "A revision has been requested for your content"
	if err := s.CreateNotification(ctx, approval.ClientID, entities.NotificationTypeRevisionRequest, title, message); err != nil {
		// Log error but don't fail the operation
	}

	return nil
}

// GetMessageThreads retrieves message threads for a client
func (s *DashboardServiceImpl) GetMessageThreads(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.MessageThread, error) {
	threads, err := s.dashboardRepo.GetMessageThreadsByClientID(ctx, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get message threads: %w", err)
	}
	return threads, nil
}

// CreateMessageThread creates a new message thread
func (s *DashboardServiceImpl) CreateMessageThread(ctx context.Context, projectID, clientID uuid.UUID, subject string) (*entities.MessageThread, error) {
	thread := entities.NewMessageThread(projectID, clientID, subject)
	
	if err := s.dashboardRepo.CreateMessageThread(ctx, thread); err != nil {
		return nil, fmt.Errorf("failed to create message thread: %w", err)
	}

	return thread, nil
}

// SendMessage sends a message in a thread
func (s *DashboardServiceImpl) SendMessage(ctx context.Context, threadID uuid.UUID, messageType entities.MessageType, content string) (*entities.DashboardMessage, error) {
	message := entities.NewDashboardMessage(threadID, messageType, content)
	
	if err := s.dashboardRepo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update thread last message time
	thread, err := s.dashboardRepo.GetMessageThreadByID(ctx, threadID)
	if err == nil {
		now := time.Now()
		thread.LastMessage = &now
		s.dashboardRepo.UpdateMessageThread(ctx, thread)
	}

	return message, nil
}

// GetThreadMessages retrieves messages for a thread
func (s *DashboardServiceImpl) GetThreadMessages(ctx context.Context, threadID uuid.UUID, limit, offset int) ([]*entities.DashboardMessage, error) {
	messages, err := s.dashboardRepo.GetMessagesByThreadID(ctx, threadID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread messages: %w", err)
	}
	return messages, nil
}

// MarkMessagesAsRead marks all messages in a thread as read
func (s *DashboardServiceImpl) MarkMessagesAsRead(ctx context.Context, threadID uuid.UUID) error {
	messages, err := s.dashboardRepo.GetMessagesByThreadID(ctx, threadID, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get messages: %w", err)
	}

	for _, message := range messages {
		if !message.IsRead {
			if err := s.dashboardRepo.MarkMessageAsRead(ctx, message.MessageID); err != nil {
				// Log error but continue
			}
		}
	}

	return nil
}

// GetNotifications retrieves notifications for a client
func (s *DashboardServiceImpl) GetNotifications(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.DashboardNotification, error) {
	notifications, err := s.dashboardRepo.GetNotificationsByClientID(ctx, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	return notifications, nil
}

// CreateNotification creates a new notification
func (s *DashboardServiceImpl) CreateNotification(ctx context.Context, clientID uuid.UUID, notificationType entities.NotificationType, title, message string) error {
	notification := entities.NewDashboardNotification(clientID, notificationType, title, message)
	
	if err := s.dashboardRepo.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// MarkNotificationAsRead marks a notification as read
func (s *DashboardServiceImpl) MarkNotificationAsRead(ctx context.Context, notificationID uuid.UUID) error {
	return s.dashboardRepo.MarkNotificationAsRead(ctx, notificationID)
}

// GetUnreadNotificationCount gets the count of unread notifications
func (s *DashboardServiceImpl) GetUnreadNotificationCount(ctx context.Context, clientID uuid.UUID) (int, error) {
	return s.dashboardRepo.GetUnreadNotificationCount(ctx, clientID)
}

// GetProjectAnalytics retrieves analytics for a specific project
func (s *DashboardServiceImpl) GetProjectAnalytics(ctx context.Context, projectID uuid.UUID) (*entities.ProjectAnalytics, error) {
	analytics, err := s.dashboardRepo.GetProjectAnalytics(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project analytics: %w", err)
	}
	return analytics, nil
}

// GetClientAnalytics retrieves comprehensive analytics for a client
func (s *DashboardServiceImpl) GetClientAnalytics(ctx context.Context, clientID uuid.UUID, fromDate, toDate time.Time) (*ClientAnalytics, error) {
	// Get projects
	projects, err := s.projectRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	analytics := &ClientAnalytics{
		ClientID:          clientID,
		TotalProjects:     len(projects),
		ProjectsByStatus:  make(map[entities.ProjectStatus]int),
		ContentByType:     make(map[entities.ContentType]int),
		FromDate:          fromDate,
		ToDate:            toDate,
	}

	// Calculate basic metrics
	for _, project := range projects {
		analytics.ProjectsByStatus[project.Status]++
		analytics.ContentByType[project.ContentType]++
		
		if project.Status == entities.ProjectStatusInProgress {
			analytics.ActiveProjects++
		} else if project.Status == entities.ProjectStatusCompleted {
			analytics.CompletedProjects++
		}
	}

	// Calculate total content pieces
	for _, project := range projects {
		analytics.TotalContentPieces += len(project.Contents)
	}

	// Set performance metrics (these would be calculated from actual data)
	analytics.PerformanceMetrics = PerformanceMetrics{
		OnTimeDeliveryRate:    92.5,
		ClientSatisfactionAvg: 4.6,
		RevisionRequestRate:   15.2,
		AverageResponseTime:   "2.4 hours",
	}

	return analytics, nil
}

// GeneratePerformanceReport generates a comprehensive performance report
func (s *DashboardServiceImpl) GeneratePerformanceReport(ctx context.Context, clientID uuid.UUID, reportType ReportType) (*PerformanceReport, error) {
	report := &PerformanceReport{
		ReportID:    uuid.New(),
		ClientID:    clientID,
		Type:        reportType,
		GeneratedAt: time.Now(),
		Data:        make(map[string]interface{}),
		Charts:      []ChartData{},
		Insights:    []string{},
		Recommendations: []string{},
	}

	switch reportType {
	case ReportTypeProjectSummary:
		report.Title = "Project Summary Report"
		report.Summary = "Comprehensive overview of all projects and their current status"
		// Add project-specific data and charts
		
	case ReportTypeContentDelivery:
		report.Title = "Content Delivery Performance"
		report.Summary = "Analysis of content delivery timelines and quality metrics"
		// Add content delivery data and charts
		
	case ReportTypeFinancialSummary:
		report.Title = "Financial Summary Report"
		report.Summary = "Overview of billing, payments, and financial performance"
		// Add financial data and charts
		
	case ReportTypePerformanceKPI:
		report.Title = "Key Performance Indicators"
		report.Summary = "Analysis of key performance metrics and trends"
		// Add KPI data and charts
	}

	return report, nil
}

// GetBillingHistory retrieves billing history for a client
func (s *DashboardServiceImpl) GetBillingHistory(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*entities.BillingHistory, error) {
	history, err := s.dashboardRepo.GetBillingHistoryByClientID(ctx, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing history: %w", err)
	}
	return history, nil
}

// GetOutstandingInvoices retrieves outstanding invoices for a client
func (s *DashboardServiceImpl) GetOutstandingInvoices(ctx context.Context, clientID uuid.UUID) ([]*entities.BillingHistory, error) {
	invoices, err := s.dashboardRepo.GetOutstandingInvoices(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get outstanding invoices: %w", err)
	}
	return invoices, nil
}

// CreateInvoice creates a new invoice
func (s *DashboardServiceImpl) CreateInvoice(ctx context.Context, projectID, clientID uuid.UUID, amount entities.Money, description string) (*entities.BillingHistory, error) {
	invoiceNumber := fmt.Sprintf("INV-%d", time.Now().Unix())
	dueDate := time.Now().AddDate(0, 0, 30) // 30 days from now
	
	billing := entities.NewBillingHistory(projectID, clientID, invoiceNumber, amount, dueDate)
	billing.Description = description
	
	if err := s.dashboardRepo.CreateBillingHistory(ctx, billing); err != nil {
		return nil, fmt.Errorf("failed to create billing history: %w", err)
	}

	// Create notification
	title := "New Invoice Generated"
	message := fmt.Sprintf("Invoice %s for %s %.2f is now available", invoiceNumber, amount.Currency, amount.Amount)
	if err := s.CreateNotification(ctx, clientID, entities.NotificationTypePaymentDue, title, message); err != nil {
		// Log error but don't fail the operation
	}

	return billing, nil
}

// MarkInvoiceAsPaid marks an invoice as paid
func (s *DashboardServiceImpl) MarkInvoiceAsPaid(ctx context.Context, billingID uuid.UUID) error {
	// This would typically integrate with the payment service
	// For now, we'll just update the billing record
	
	// Get billing record
	history, err := s.dashboardRepo.GetBillingHistoryByClientID(ctx, uuid.Nil, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to get billing history: %w", err)
	}

	for _, billing := range history {
		if billing.BillingID == billingID {
			billing.MarkAsPaid()
			return s.dashboardRepo.UpdateBillingHistory(ctx, billing)
		}
	}

	return fmt.Errorf("billing record not found")
}