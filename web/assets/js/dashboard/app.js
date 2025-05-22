// Dashboard Main Application Module

import { core } from './core.js';
import { api } from './api.js';
import { NotificationManager } from './notifications.js';
import { ProjectManager } from './projects.js';
import { ContentManager } from './content.js';
import { MessageManager } from './messages.js';
import { BillingManager } from './billing.js';
import { AnalyticsManager } from './analytics.js';

class DashboardApp {
    constructor() {
        this.currentView = 'overview';
        this.managers = {};
        this.isInitialized = false;
        this.refreshInterval = null;
        this.refreshRate = 30000; // 30 seconds
    }

    async init() {
        if (this.isInitialized) return;

        try {
            // Initialize core
            core.init();
            
            // Setup event listeners
            this.setupEventListeners();
            
            // Initialize managers
            await this.initializeManagers();
            
            // Setup navigation
            this.setupNavigation();
            
            // Setup auto-refresh
            this.setupAutoRefresh();
            
            // Load initial data
            await this.loadInitialData();
            
            // Hide loading screen
            this.hideLoadingScreen();
            
            this.isInitialized = true;
            core.handleSuccess('Dashboard loaded successfully');
            
        } catch (error) {
            core.handleError(error, 'Dashboard Initialization');
            this.showError('Failed to initialize dashboard. Please refresh the page.');
        }
    }

    async initializeManagers() {
        // Initialize all feature managers
        this.managers.notifications = new NotificationManager();
        this.managers.projects = new ProjectManager();
        this.managers.content = new ContentManager();
        this.managers.messages = new MessageManager();
        this.managers.billing = new BillingManager();
        this.managers.analytics = new AnalyticsManager();

        // Initialize each manager
        const initPromises = Object.values(this.managers).map(manager => 
            manager.init ? manager.init() : Promise.resolve()
        );

        await Promise.all(initPromises);
    }

    setupEventListeners() {
        // Sidebar toggle for mobile
        const sidebarToggle = core.$('#sidebarToggle');
        const sidebar = core.$('#sidebar');
        
        if (sidebarToggle && sidebar) {
            sidebarToggle.addEventListener('click', () => {
                sidebar.classList.toggle('show');
            });
        }

        // Close sidebar when clicking outside on mobile
        document.addEventListener('click', (e) => {
            if (window.innerWidth <= 768) {
                const sidebar = core.$('#sidebar');
                const sidebarToggle = core.$('#sidebarToggle');
                
                if (sidebar && sidebar.classList.contains('show') && 
                    !sidebar.contains(e.target) && 
                    !sidebarToggle.contains(e.target)) {
                    sidebar.classList.remove('show');
                }
            }
        });

        // Refresh button
        const refreshBtn = core.$('#refreshBtn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => this.refreshData());
        }

        // Search focus shortcut
        core.on('search-focus', () => {
            const searchInput = core.$('.search-input:not([style*="display: none"])');
            if (searchInput) {
                searchInput.focus();
            }
        });

        // Escape key handling
        core.on('escape-pressed', () => {
            this.closeAllModals();
        });

        // Window resize handling
        window.addEventListener('resize', core.debounce(() => {
            this.handleResize();
        }, 250));

        // Online/offline handling
        core.on('online', () => {
            this.resumeAutoRefresh();
            this.refreshData();
        });

        core.on('offline', () => {
            this.pauseAutoRefresh();
        });
    }

    setupNavigation() {
        const navLinks = core.$$('.nav-link');
        
        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const view = link.dataset.view;
                if (view) {
                    this.switchView(view);
                }
            });
        });

        // Handle browser back/forward
        window.addEventListener('popstate', (e) => {
            const view = e.state?.view || 'overview';
            this.switchView(view, false);
        });

        // Set initial view from URL hash
        const hash = window.location.hash.substring(1);
        if (hash && this.isValidView(hash)) {
            this.switchView(hash, false);
        }
    }

    setupAutoRefresh() {
        this.refreshInterval = setInterval(() => {
            if (document.visibilityState === 'visible') {
                this.refreshData(true); // Silent refresh
            }
        }, this.refreshRate);

        // Pause refresh when tab is not visible
        document.addEventListener('visibilitychange', () => {
            if (document.visibilityState === 'visible') {
                this.resumeAutoRefresh();
            } else {
                this.pauseAutoRefresh();
            }
        });
    }

    async loadInitialData() {
        try {
            // Load dashboard summary
            await this.loadDashboardSummary();
            
            // Load current view data
            await this.loadCurrentViewData();
            
        } catch (error) {
            console.error('Error loading initial data:', error);
            // Continue with cached or mock data
            this.loadMockData();
        }
    }

    async loadDashboardSummary() {
        try {
            const summary = await api.getDashboardSummary(core.currentClientId);
            this.updateDashboardSummary(summary);
        } catch (error) {
            // Use mock data for development
            const mockSummary = api.getMockDashboardSummary();
            this.updateDashboardSummary(mockSummary);
        }
    }

    updateDashboardSummary(summary) {
        // Update header badges
        this.updateBadge('activeProjectsBadge', summary.activeProjects);
        this.updateBadge('pendingApprovalsBadge', summary.pendingApprovals);
        this.updateBadge('unreadMessagesBadge', summary.unreadMessages);
        this.updateBadge('notificationsBadge', summary.unreadNotifications);

        // Update overview stats
        this.updateStat('activeProjectsCount', summary.activeProjects);
        this.updateStat('completedProjectsCount', summary.completedProjects);
        this.updateStat('pendingApprovalsCount', summary.pendingApprovals);
        this.updateStat('outstandingBalance', core.formatCurrency(summary.outstandingBalance.amount));

        // Update recent activity
        this.updateRecentActivity(summary.recentActivity || []);
        
        // Update upcoming deadlines
        this.updateUpcomingDeadlines(summary.upcomingDeadlines || []);
    }

    updateBadge(elementId, count) {
        const badge = core.$(`#${elementId}`);
        if (badge) {
            badge.textContent = count;
            badge.style.display = count > 0 ? 'flex' : 'none';
        }
    }

    updateStat(elementId, value) {
        const element = core.$(`#${elementId}`);
        if (element) {
            element.textContent = value;
        }
    }

    updateRecentActivity(activities) {
        const container = core.$('#recentActivityList');
        if (!container) return;

        container.innerHTML = '';

        if (activities.length === 0) {
            container.innerHTML = '<p class="no-data">No recent activity</p>';
            return;
        }

        activities.forEach(activity => {
            const activityElement = core.createElement('div', { className: 'activity-item' }, `
                <div class="activity-icon">${this.getActivityIcon(activity.type)}</div>
                <div class="activity-content">
                    <div class="activity-description">${activity.description}</div>
                    <div class="activity-time">${core.getRelativeTime(activity.timestamp)}</div>
                </div>
            `);

            if (activity.projectId) {
                activityElement.style.cursor = 'pointer';
                activityElement.addEventListener('click', () => {
                    this.managers.projects.showProjectDetails(activity.projectId);
                });
            }

            container.appendChild(activityElement);
        });
    }

    updateUpcomingDeadlines(deadlines) {
        const container = core.$('#upcomingDeadlinesList');
        if (!container) return;

        container.innerHTML = '';

        if (deadlines.length === 0) {
            container.innerHTML = '<p class="no-data">No upcoming deadlines</p>';
            return;
        }

        deadlines.forEach(deadline => {
            const statusClass = deadline.isOverdue ? 'overdue' : 
                               deadline.daysLeft <= 1 ? 'urgent' : 'upcoming';
            
            const deadlineElement = core.createElement('div', { className: 'deadline-item' }, `
                <div class="deadline-content">
                    <div class="deadline-project">${deadline.projectName}</div>
                    <div class="deadline-date">${core.formatDate(deadline.deadline)}</div>
                </div>
                <div class="deadline-status ${statusClass}">
                    ${deadline.isOverdue ? 'Overdue' : `${deadline.daysLeft} days`}
                </div>
            `);

            deadlineElement.style.cursor = 'pointer';
            deadlineElement.addEventListener('click', () => {
                this.managers.projects.showProjectDetails(deadline.projectId);
            });

            container.appendChild(deadlineElement);
        });
    }

    getActivityIcon(type) {
        const icons = {
            content_delivered: '📄',
            payment_received: '💳',
            project_started: '🚀',
            project_completed: '✅',
            message_received: '💬',
            approval_requested: '👀',
            default: '📋'
        };
        return icons[type] || icons.default;
    }

    async loadCurrentViewData() {
        if (this.managers[this.currentView]?.loadData) {
            await this.managers[this.currentView].loadData();
        }
    }

    loadMockData() {
        // Load mock data for development
        console.log('Loading mock data for development');
        const mockSummary = api.getMockDashboardSummary();
        this.updateDashboardSummary(mockSummary);
    }

    switchView(viewName, updateHistory = true) {
        if (!this.isValidView(viewName) || viewName === this.currentView) return;

        // Update navigation
        core.$$('.nav-link').forEach(link => {
            link.classList.toggle('active', link.dataset.view === viewName);
        });

        // Update views
        core.$$('.view').forEach(view => {
            view.classList.toggle('active', view.id === `${viewName}View`);
        });

        // Update page title and subtitle
        this.updatePageHeader(viewName);

        // Update URL
        if (updateHistory) {
            const url = new URL(window.location);
            url.hash = viewName;
            history.pushState({ view: viewName }, '', url);
        }

        // Load view data
        this.currentView = viewName;
        this.loadCurrentViewData();

        // Close sidebar on mobile
        if (window.innerWidth <= 768) {
            const sidebar = core.$('#sidebar');
            if (sidebar) {
                sidebar.classList.remove('show');
            }
        }
    }

    updatePageHeader(viewName) {
        const titles = {
            overview: { title: 'Dashboard Overview', subtitle: "Welcome back! Here's what's happening with your projects." },
            projects: { title: 'Projects', subtitle: 'Manage and track your content creation projects.' },
            content: { title: 'Content Approvals', subtitle: 'Review and approve your content before delivery.' },
            messages: { title: 'Messages', subtitle: 'Communicate with your content creation team.' },
            billing: { title: 'Billing & Payments', subtitle: 'View your billing history and outstanding invoices.' },
            analytics: { title: 'Analytics & Reports', subtitle: 'Track your content performance and business metrics.' }
        };

        const config = titles[viewName] || titles.overview;
        
        const titleElement = core.$('#pageTitle');
        const subtitleElement = core.$('#pageSubtitle');
        
        if (titleElement) titleElement.textContent = config.title;
        if (subtitleElement) subtitleElement.textContent = config.subtitle;
    }

    isValidView(viewName) {
        const validViews = ['overview', 'projects', 'content', 'messages', 'billing', 'analytics'];
        return validViews.includes(viewName);
    }

    async refreshData(silent = false) {
        const refreshBtn = core.$('#refreshBtn');
        
        if (!silent && refreshBtn) {
            refreshBtn.classList.add('loading');
        }

        try {
            // Clear cache for current client
            core.clearCache(core.currentClientId);
            
            // Reload dashboard summary
            await this.loadDashboardSummary();
            
            // Reload current view data
            await this.loadCurrentViewData();
            
            if (!silent) {
                core.handleSuccess('Data refreshed successfully');
            }
            
        } catch (error) {
            if (!silent) {
                core.handleError(error, 'Data Refresh');
            }
        } finally {
            if (refreshBtn) {
                refreshBtn.classList.remove('loading');
            }
        }
    }

    pauseAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    resumeAutoRefresh() {
        if (!this.refreshInterval) {
            this.setupAutoRefresh();
        }
    }

    handleResize() {
        // Handle responsive layout changes
        if (window.innerWidth > 768) {
            const sidebar = core.$('#sidebar');
            if (sidebar) {
                sidebar.classList.remove('show');
            }
        }
    }

    closeAllModals() {
        core.$$('.modal.show').forEach(modal => {
            modal.classList.remove('show');
        });
    }

    hideLoadingScreen() {
        const loadingScreen = core.$('#loadingScreen');
        if (loadingScreen) {
            loadingScreen.classList.add('hidden');
            setTimeout(() => {
                loadingScreen.style.display = 'none';
            }, 300);
        }
    }

    showError(message) {
        const loadingScreen = core.$('#loadingScreen');
        if (loadingScreen) {
            loadingScreen.innerHTML = `
                <div class="error-state">
                    <div class="error-icon">⚠️</div>
                    <h2>Oops! Something went wrong</h2>
                    <p>${message}</p>
                    <button onclick="window.location.reload()" class="retry-btn">
                        Retry
                    </button>
                </div>
            `;
        }
    }

    // Public API
    getCurrentView() {
        return this.currentView;
    }

    getManager(name) {
        return this.managers[name];
    }

    // Cleanup
    destroy() {
        this.pauseAutoRefresh();
        
        Object.values(this.managers).forEach(manager => {
            if (manager.destroy) {
                manager.destroy();
            }
        });
        
        this.isInitialized = false;
    }
}

// Initialize dashboard when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.dashboard = new DashboardApp();
    window.dashboard.init().catch(error => {
        console.error('Failed to initialize dashboard:', error);
    });
});

// Export for global access
export { DashboardApp };