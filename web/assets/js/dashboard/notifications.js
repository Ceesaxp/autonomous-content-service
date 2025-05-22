// Dashboard Notifications Module

import { core } from './core.js';
import { api } from './api.js';

export class NotificationManager {
    constructor() {
        this.notifications = [];
        this.isDropdownOpen = false;
        this.unreadCount = 0;
        this.pollInterval = null;
        this.pollRate = 30000; // 30 seconds
    }

    async init() {
        this.setupEventListeners();
        this.startPolling();
        await this.loadNotifications();
    }

    setupEventListeners() {
        // Notifications button
        const notificationsBtn = core.$('#notificationsBtn');
        const notificationsDropdown = core.$('#notificationsDropdown');
        
        if (notificationsBtn && notificationsDropdown) {
            notificationsBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.toggleDropdown();
            });

            // Close dropdown when clicking outside
            document.addEventListener('click', (e) => {
                if (!notificationsDropdown.contains(e.target) && 
                    !notificationsBtn.contains(e.target)) {
                    this.closeDropdown();
                }
            });
        }

        // Mark all read button
        const markAllReadBtn = core.$('#markAllReadBtn');
        if (markAllReadBtn) {
            markAllReadBtn.addEventListener('click', () => {
                this.markAllAsRead();
            });
        }

        // Listen for escape key
        core.on('escape-pressed', () => {
            this.closeDropdown();
        });
    }

    async loadNotifications() {
        try {
            this.notifications = await api.getNotifications(core.currentClientId, 20, 0);
            this.updateUI();
        } catch (error) {
            console.error('Error loading notifications:', error);
            // Use mock data for development
            this.notifications = this.getMockNotifications();
            this.updateUI();
        }
    }

    updateUI() {
        this.updateUnreadCount();
        this.renderNotifications();
    }

    updateUnreadCount() {
        this.unreadCount = this.notifications.filter(n => !n.isRead).length;
        
        const badge = core.$('#notificationsBadge');
        if (badge) {
            badge.textContent = this.unreadCount;
            badge.classList.toggle('hidden', this.unreadCount === 0);
        }

        // Update mark all read button
        const markAllReadBtn = core.$('#markAllReadBtn');
        if (markAllReadBtn) {
            markAllReadBtn.style.display = this.unreadCount > 0 ? 'block' : 'none';
        }
    }

    renderNotifications() {
        const container = core.$('#notificationsList');
        if (!container) return;

        container.innerHTML = '';

        if (this.notifications.length === 0) {
            container.innerHTML = `
                <div class="no-notifications">
                    <p>No notifications yet</p>
                </div>
            `;
            return;
        }

        this.notifications.forEach(notification => {
            const notificationElement = this.createNotificationElement(notification);
            container.appendChild(notificationElement);
        });
    }

    createNotificationElement(notification) {
        const element = core.createElement('div', {
            className: `notification-item ${notification.isRead ? '' : 'unread'}`,
            'data-notification-id': notification.notificationId
        });

        const timeAgo = core.getRelativeTime(notification.createdAt);
        const priorityIcon = this.getPriorityIcon(notification.priority);

        element.innerHTML = `
            <div class="notification-content">
                <div class="notification-header">
                    <span class="notification-priority">${priorityIcon}</span>
                    <div class="notification-title">${notification.title}</div>
                </div>
                <div class="notification-message">${notification.message}</div>
                <div class="notification-time">${timeAgo}</div>
            </div>
        `;

        // Add click handler
        element.addEventListener('click', () => {
            this.handleNotificationClick(notification);
        });

        return element;
    }

    getPriorityIcon(priority) {
        const icons = {
            'High': '🔴',
            'Medium': '🟡',
            'Low': '🟢'
        };
        return icons[priority] || icons['Medium'];
    }

    async handleNotificationClick(notification) {
        // Mark as read if unread
        if (!notification.isRead) {
            await this.markAsRead(notification.notificationId);
        }

        // Handle action URL
        if (notification.actionUrl) {
            if (notification.actionUrl.startsWith('#')) {
                // Internal navigation
                const view = notification.actionUrl.substring(1);
                core.emit('navigate', { view });
            } else {
                // External URL
                window.open(notification.actionUrl, '_blank');
            }
        }

        // Handle specific notification types
        this.handleNotificationType(notification);
        
        // Close dropdown
        this.closeDropdown();
    }

    handleNotificationType(notification) {
        switch (notification.type) {
            case 'ContentReady':
                core.emit('navigate', { view: 'content' });
                break;
            case 'RevisionRequest':
                core.emit('navigate', { view: 'content' });
                break;
            case 'PaymentDue':
                core.emit('navigate', { view: 'billing' });
                break;
            case 'ProjectUpdate':
                if (notification.projectId) {
                    core.emit('show-project-details', { projectId: notification.projectId });
                }
                break;
            case 'Message':
                core.emit('navigate', { view: 'messages' });
                break;
            case 'DeadlineAlert':
                core.emit('navigate', { view: 'projects' });
                break;
        }
    }

    async markAsRead(notificationId) {
        try {
            await api.markNotificationAsRead(notificationId);
            
            // Update local state
            const notification = this.notifications.find(n => n.notificationId === notificationId);
            if (notification) {
                notification.isRead = true;
                notification.readAt = new Date().toISOString();
            }
            
            this.updateUI();
            
        } catch (error) {
            core.handleError(error, 'Mark Notification Read');
        }
    }

    async markAllAsRead() {
        const unreadNotifications = this.notifications.filter(n => !n.isRead);
        
        if (unreadNotifications.length === 0) return;

        try {
            // Mark all unread notifications as read
            const promises = unreadNotifications.map(notification => 
                api.markNotificationAsRead(notification.notificationId)
            );
            
            await Promise.all(promises);
            
            // Update local state
            unreadNotifications.forEach(notification => {
                notification.isRead = true;
                notification.readAt = new Date().toISOString();
            });
            
            this.updateUI();
            core.handleSuccess('All notifications marked as read');
            
        } catch (error) {
            core.handleError(error, 'Mark All Notifications Read');
        }
    }

    toggleDropdown() {
        const dropdown = core.$('#notificationsDropdown');
        if (!dropdown) return;

        this.isDropdownOpen = !this.isDropdownOpen;
        dropdown.classList.toggle('show', this.isDropdownOpen);

        if (this.isDropdownOpen) {
            // Refresh notifications when opening
            this.loadNotifications();
        }
    }

    closeDropdown() {
        const dropdown = core.$('#notificationsDropdown');
        if (dropdown) {
            dropdown.classList.remove('show');
            this.isDropdownOpen = false;
        }
    }

    startPolling() {
        this.pollInterval = setInterval(() => {
            if (document.visibilityState === 'visible') {
                this.loadNotifications();
            }
        }, this.pollRate);
    }

    stopPolling() {
        if (this.pollInterval) {
            clearInterval(this.pollInterval);
            this.pollInterval = null;
        }
    }

    // Create a new notification (for real-time updates)
    addNotification(notification) {
        this.notifications.unshift(notification);
        
        // Limit to last 50 notifications
        if (this.notifications.length > 50) {
            this.notifications = this.notifications.slice(0, 50);
        }
        
        this.updateUI();
        
        // Show toast for new notifications
        if (!notification.isRead) {
            core.showToast(notification.title, 'info', 3000);
        }
    }

    getMockNotifications() {
        return [
            {
                notificationId: 'notif-1',
                clientId: core.currentClientId,
                projectId: 'proj-123',
                type: 'ContentReady',
                priority: 'High',
                title: 'Content Ready for Review',
                message: 'Your blog post "SEO Best Practices" is ready for approval.',
                actionUrl: '#content',
                isRead: false,
                createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString()
            },
            {
                notificationId: 'notif-2',
                clientId: core.currentClientId,
                type: 'PaymentDue',
                priority: 'Medium',
                title: 'Invoice Due Soon',
                message: 'Invoice #INV-001 for $1,250 is due in 3 days.',
                actionUrl: '#billing',
                isRead: false,
                createdAt: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString()
            },
            {
                notificationId: 'notif-3',
                clientId: core.currentClientId,
                projectId: 'proj-124',
                type: 'ProjectUpdate',
                priority: 'Low',
                title: 'Project Status Updated',
                message: 'Social Media Strategy project moved to review stage.',
                isRead: true,
                createdAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
                readAt: new Date(Date.now() - 12 * 60 * 60 * 1000).toISOString()
            },
            {
                notificationId: 'notif-4',
                clientId: core.currentClientId,
                type: 'Message',
                priority: 'Medium',
                title: 'New Message',
                message: 'You have a new message about your content requirements.',
                actionUrl: '#messages',
                isRead: false,
                createdAt: new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString()
            },
            {
                notificationId: 'notif-5',
                clientId: core.currentClientId,
                projectId: 'proj-123',
                type: 'DeadlineAlert',
                priority: 'High',
                title: 'Deadline Approaching',
                message: 'SEO Content Package deadline is in 3 days.',
                actionUrl: '#projects',
                isRead: true,
                createdAt: new Date(Date.now() - 48 * 60 * 60 * 1000).toISOString(),
                readAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
            }
        ];
    }

    destroy() {
        this.stopPolling();
    }
}