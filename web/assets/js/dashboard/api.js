// Dashboard API Module - Handles all API communications

import { core } from './core.js';

export class DashboardAPI {
    constructor() {
        this.baseUrl = core.apiBaseUrl;
        this.defaultHeaders = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
    }

    // Generic HTTP request method
    async request(endpoint, options = {}) {
        const url = `${this.baseUrl}${endpoint}`;
        const config = {
            headers: { ...this.defaultHeaders, ...options.headers },
            ...options
        };

        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                return await response.json();
            }
            
            return await response.text();
        } catch (error) {
            console.error(`API Error [${endpoint}]:`, error);
            throw error;
        }
    }

    // HTTP method helpers
    async get(endpoint, params = {}) {
        const url = new URL(`${this.baseUrl}${endpoint}`);
        Object.entries(params).forEach(([key, value]) => {
            if (value !== null && value !== undefined) {
                url.searchParams.append(key, value);
            }
        });

        return this.request(url.pathname + url.search, { method: 'GET' });
    }

    async post(endpoint, data = null) {
        return this.request(endpoint, {
            method: 'POST',
            body: data ? JSON.stringify(data) : null
        });
    }

    async put(endpoint, data = null) {
        return this.request(endpoint, {
            method: 'PUT',
            body: data ? JSON.stringify(data) : null
        });
    }

    async delete(endpoint) {
        return this.request(endpoint, { method: 'DELETE' });
    }

    // Dashboard Summary API
    async getDashboardSummary(clientId) {
        const cacheKey = `dashboard-summary-${clientId}`;
        const cached = core.getCache(cacheKey);
        if (cached) return cached;

        const data = await this.get(`/dashboard/summary/${clientId}`);
        core.setCache(cacheKey, data, 2 * 60 * 1000); // 2 minutes cache
        return data;
    }

    async refreshDashboard(clientId) {
        core.clearCache(`dashboard-${clientId}`);
        return this.post(`/dashboard/refresh/${clientId}`);
    }

    // Projects API
    async getProjectsOverview(clientId) {
        const cacheKey = `projects-overview-${clientId}`;
        const cached = core.getCache(cacheKey);
        if (cached) return cached;

        const data = await this.get(`/dashboard/projects/${clientId}`);
        core.setCache(cacheKey, data);
        return data;
    }

    async getProjectDetails(projectId) {
        const cacheKey = `project-details-${projectId}`;
        const cached = core.getCache(cacheKey);
        if (cached) return cached;

        const data = await this.get(`/dashboard/projects/details/${projectId}`);
        core.setCache(cacheKey, data);
        return data;
    }

    async updateProjectStatus(projectId, status) {
        const result = await this.put(`/dashboard/projects/${projectId}/status`, { status });
        
        // Clear related cache
        core.clearCache('projects-overview');
        core.clearCache(`project-details-${projectId}`);
        core.clearCache('dashboard-summary');
        
        return result;
    }

    // Content Approvals API
    async getContentApprovals(clientId, limit = 20, offset = 0) {
        const cacheKey = `content-approvals-${clientId}-${limit}-${offset}`;
        const cached = core.getCache(cacheKey);
        if (cached) return cached;

        const data = await this.get(`/dashboard/approvals/${clientId}`, { limit, offset });
        core.setCache(cacheKey, data);
        return data;
    }

    async approveContent(approvalId, feedback = '') {
        const result = await this.put(`/dashboard/approvals/${approvalId}/approve`, { feedback });
        
        // Clear related cache
        core.clearCache('content-approvals');
        core.clearCache('dashboard-summary');
        
        return result;
    }

    async rejectContent(approvalId, feedback) {
        const result = await this.put(`/dashboard/approvals/${approvalId}/reject`, { feedback });
        
        // Clear related cache
        core.clearCache('content-approvals');
        core.clearCache('dashboard-summary');
        
        return result;
    }

    async requestContentRevision(approvalId, feedback) {
        const result = await this.put(`/dashboard/approvals/${approvalId}/revision`, { feedback });
        
        // Clear related cache
        core.clearCache('content-approvals');
        core.clearCache('dashboard-summary');
        
        return result;
    }

    // Messages API
    async getMessageThreads(clientId, limit = 20, offset = 0) {
        const cacheKey = `message-threads-${clientId}-${limit}-${offset}`;
        const cached = core.getCache(cacheKey);
        if (cached) return cached;

        const data = await this.get(`/dashboard/messages/${clientId}`, { limit, offset });
        core.setCache(cacheKey, data, 30 * 1000); // 30 seconds cache for messages
        return data;
    }

    async createMessageThread(projectId, clientId, subject) {
        const result = await this.post('/dashboard/messages/threads', {
            projectId,
            clientId,
            subject
        });
        
        // Clear related cache
        core.clearCache('message-threads');
        
        return result;
    }

    async getThreadMessages(threadId, limit = 50, offset = 0) {
        const cacheKey = `thread-messages-${threadId}-${limit}-${offset}`;
        const cached = core.getCache(cacheKey);
        if (cached) return cached;

        const data = await this.get(`/dashboard/messages/${threadId}/messages`, { limit, offset });
        core.setCache(cacheKey, data, 10 * 1000); // 10 seconds cache for messages
        return data;
    }

    async sendMessage(threadId, type, content) {
        const result = await this.post(`/dashboard/messages/${threadId}/send`, {
            type,
            content
        });
        
        // Clear related cache
        core.clearCache(`thread-messages-${threadId}`);
        core.clearCache('message-threads');
        
        return result;
    }

    async markMessagesAsRead(threadId) {
        const result = await this.put(`/dashboard/messages/${threadId}/read`);
        
        // Clear related cache
        core.clearCache(`thread-messages-${threadId}`);
        core.clearCache('message-threads');
        core.clearCache('dashboard-summary');
        
        return result;
    }

    // Notifications API
    async getNotifications(clientId, limit = 20, offset = 0) {
        const cacheKey = `notifications-${clientId}-${limit}-${offset}`;
        const cached = core.getCache(cacheKey);
        if (cached) return cached;

        const data = await this.get(`/dashboard/notifications/${clientId}`, { limit, offset });
        core.setCache(cacheKey, data, 30 * 1000); // 30 seconds cache
        return data;
    }

    async markNotificationAsRead(notificationId) {
        const result = await this.put(`/dashboard/notifications/${notificationId}/read`);
        
        // Clear related cache
        core.clearCache('notifications');
        core.clearCache('dashboard-summary');
        
        return result;
    }

    async getUnreadNotificationCount(clientId) {
        return this.get(`/dashboard/notifications/${clientId}/count`);
    }

    // Analytics API
    async getClientAnalytics(clientId, fromDate, toDate) {
        const cacheKey = `analytics-${clientId}-${fromDate}-${toDate}`;
        const cached = core.getCache(cacheKey);
        if (cached) return cached;

        const data = await this.get(`/dashboard/analytics/${clientId}`, {
            from: fromDate,
            to: toDate
        });
        core.setCache(cacheKey, data, 10 * 60 * 1000); // 10 minutes cache
        return data;
    }

    async getProjectAnalytics(projectId) {
        const cacheKey = `project-analytics-${projectId}`;
        const cached = core.getCache(cacheKey);
        if (cached) return cached;

        const data = await this.get(`/dashboard/analytics/project/${projectId}`);
        core.setCache(cacheKey, data);
        return data;
    }

    async generateReport(clientId, reportType) {
        return this.post(`/dashboard/reports/${clientId}`, { type: reportType });
    }

    // Billing API
    async getBillingHistory(clientId, limit = 20, offset = 0) {
        const cacheKey = `billing-history-${clientId}-${limit}-${offset}`;
        const cached = core.getCache(cacheKey);
        if (cached) return cached;

        const data = await this.get(`/dashboard/billing/${clientId}`, { limit, offset });
        core.setCache(cacheKey, data);
        return data;
    }

    async getOutstandingInvoices(clientId) {
        const cacheKey = `outstanding-invoices-${clientId}`;
        const cached = core.getCache(cacheKey);
        if (cached) return cached;

        const data = await this.get(`/dashboard/billing/${clientId}/outstanding`);
        core.setCache(cacheKey, data);
        return data;
    }

    // Health Check
    async healthCheck() {
        try {
            const response = await this.get('/health');
            return { status: 'healthy', ...response };
        } catch (error) {
            return { status: 'unhealthy', error: error.message };
        }
    }

    // Mock Data Methods (for development/demo purposes)
    getMockDashboardSummary() {
        return {
            clientId: core.currentClientId,
            activeProjects: 5,
            completedProjects: 12,
            pendingApprovals: 3,
            unreadNotifications: 7,
            unreadMessages: 2,
            outstandingBalance: { amount: 2450.00, currency: 'USD' },
            projectStatusBreakdown: {
                'InProgress': 5,
                'Review': 2,
                'Planning': 1,
                'Completed': 12,
                'Draft': 0
            },
            recentActivity: [
                {
                    type: 'content_delivered',
                    description: 'Blog post "SEO Best Practices" delivered',
                    projectId: 'proj-123',
                    timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString()
                },
                {
                    type: 'payment_received',
                    description: 'Payment received for Invoice #INV-001',
                    timestamp: new Date(Date.now() - 5 * 60 * 60 * 1000).toISOString()
                },
                {
                    type: 'project_started',
                    description: 'New project "Content Marketing Strategy" started',
                    projectId: 'proj-124',
                    timestamp: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
                }
            ],
            upcomingDeadlines: [
                {
                    projectId: 'proj-123',
                    projectName: 'SEO Optimization',
                    deadline: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000).toISOString(),
                    daysLeft: 3,
                    isOverdue: false
                },
                {
                    projectId: 'proj-124',
                    projectName: 'Social Media Content',
                    deadline: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
                    daysLeft: 7,
                    isOverdue: false
                }
            ],
            lastUpdated: new Date().toISOString()
        };
    }

    getMockProjectsOverview() {
        return [
            {
                projectId: 'proj-123',
                title: 'SEO Content Package',
                status: 'InProgress',
                priority: 'High',
                progress: 75,
                deadline: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
                daysRemaining: 7,
                budget: { amount: 2500.00, currency: 'USD' },
                contentCount: 8,
                pendingApprovals: 2,
                lastUpdate: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString()
            },
            {
                projectId: 'proj-124',
                title: 'Social Media Strategy',
                status: 'Review',
                priority: 'Medium',
                progress: 90,
                deadline: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000).toISOString(),
                daysRemaining: 3,
                budget: { amount: 1800.00, currency: 'USD' },
                contentCount: 12,
                pendingApprovals: 1,
                lastUpdate: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString()
            },
            {
                projectId: 'proj-125',
                title: 'Email Marketing Campaign',
                status: 'Planning',
                priority: 'Medium',
                progress: 25,
                deadline: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString(),
                daysRemaining: 14,
                budget: { amount: 1200.00, currency: 'USD' },
                contentCount: 6,
                pendingApprovals: 0,
                lastUpdate: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
            }
        ];
    }

    getMockContentApprovals() {
        return [
            {
                approvalId: 'approval-123',
                contentId: 'content-456',
                projectId: 'proj-123',
                clientId: core.currentClientId,
                status: 'Pending',
                feedback: '',
                createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
                updatedAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
                content: {
                    title: 'SEO Best Practices for 2024',
                    type: 'blog_post',
                    wordCount: 1500
                }
            },
            {
                approvalId: 'approval-124',
                contentId: 'content-457',
                projectId: 'proj-124',
                clientId: core.currentClientId,
                status: 'Pending',
                feedback: '',
                createdAt: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString(),
                updatedAt: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString(),
                content: {
                    title: 'Social Media Content Calendar Template',
                    type: 'template',
                    wordCount: 800
                }
            }
        ];
    }
}

// Export singleton instance
export const api = new DashboardAPI();