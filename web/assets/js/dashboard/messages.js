// Dashboard Messages Module

import { core } from './core.js';
import { api } from './api.js';

export class MessageManager {
    constructor() {
        this.messageThreads = [];
        this.currentThread = null;
        this.currentMessages = [];
        this.pollInterval = null;
        this.pollRate = 10000; // 10 seconds for messages
    }

    async init() {
        this.setupEventListeners();
        await this.loadData();
        this.startPolling();
    }

    setupEventListeners() {
        // New thread button
        const newThreadBtn = core.$('#newThreadBtn');
        if (newThreadBtn) {
            newThreadBtn.addEventListener('click', () => {
                this.showNewThreadModal();
            });
        }

        // Send message button
        const sendBtn = core.$('#sendMessageBtn');
        if (sendBtn) {
            sendBtn.addEventListener('click', () => {
                this.sendMessage();
            });
        }

        // Message input enter key
        const messageInput = core.$('#messageText');
        if (messageInput) {
            messageInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
                    e.preventDefault();
                    this.sendMessage();
                }
            });
        }

        // Modal handlers
        this.setupModalHandlers();

        // Mobile responsive handlers
        this.setupMobileHandlers();
    }

    setupModalHandlers() {
        const modal = core.$('#newThreadModal');
        if (!modal) return;

        // Close button
        const closeBtn = modal.querySelector('.modal-close');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => {
                this.closeModal();
            });
        }

        // Form submission
        const form = modal.querySelector('#newThreadForm');
        if (form) {
            form.addEventListener('submit', (e) => {
                e.preventDefault();
                this.createNewThread();
            });
        }

        // Cancel button
        const cancelButtons = modal.querySelectorAll('[data-modal-close]');
        cancelButtons.forEach(btn => {
            btn.addEventListener('click', () => {
                this.closeModal();
            });
        });
    }

    setupMobileHandlers() {
        // Handle mobile view switching
        if (window.innerWidth <= 768) {
            this.setupMobileView();
        }

        window.addEventListener('resize', () => {
            if (window.innerWidth <= 768) {
                this.setupMobileView();
            } else {
                this.setupDesktopView();
            }
        });
    }

    setupMobileView() {
        const messagesLayout = core.$('.messages-layout');
        if (messagesLayout) {
            messagesLayout.classList.add('mobile-view');
        }
    }

    setupDesktopView() {
        const messagesLayout = core.$('.messages-layout');
        if (messagesLayout) {
            messagesLayout.classList.remove('mobile-view');
        }
    }

    async loadData() {
        try {
            this.messageThreads = await api.getMessageThreads(core.currentClientId, 50, 0);
            this.renderThreads();
            
            // Auto-select first thread if available
            if (this.messageThreads.length > 0 && !this.currentThread) {
                this.selectThread(this.messageThreads[0]);
            }
        } catch (error) {
            console.error('Error loading message threads:', error);
            // Use mock data for development
            this.messageThreads = this.getMockMessageThreads();
            this.renderThreads();
            
            if (this.messageThreads.length > 0) {
                this.selectThread(this.messageThreads[0]);
            }
        }
    }

    renderThreads() {
        const container = core.$('#messageThreadsList');
        if (!container) return;

        container.innerHTML = '';

        if (this.messageThreads.length === 0) {
            container.innerHTML = `
                <div class="no-threads">
                    <p>No conversations yet</p>
                    <button class="btn btn-primary" onclick="this.showNewThreadModal()">Start a conversation</button>
                </div>
            `;
            return;
        }

        this.messageThreads.forEach(thread => {
            const threadElement = this.createThreadElement(thread);
            container.appendChild(threadElement);
        });
    }

    createThreadElement(thread) {
        const element = core.createElement('div', {
            className: `thread-item ${thread.threadId === this.currentThread?.threadId ? 'active' : ''}`,
            'data-thread-id': thread.threadId
        });

        const lastMessageTime = thread.lastMessage ? 
            core.getRelativeTime(thread.lastMessage) : 
            core.getRelativeTime(thread.createdAt);

        // Mock unread status and preview for development
        const hasUnread = Math.random() > 0.7; // 30% chance of unread
        const preview = this.getMockThreadPreview(thread.subject);

        element.innerHTML = `
            <div class="thread-subject">${thread.subject}</div>
            <div class="thread-preview">${preview}</div>
            <div class="thread-time">${lastMessageTime}</div>
        `;

        if (hasUnread) {
            element.classList.add('unread');
        }

        element.addEventListener('click', () => {
            this.selectThread(thread);
        });

        return element;
    }

    async selectThread(thread) {
        if (!thread) return;

        // Update UI
        this.currentThread = thread;
        this.updateActiveThread();
        
        // Show conversation area
        this.showConversationArea();
        
        // Load messages
        await this.loadThreadMessages(thread.threadId);
        
        // Mark messages as read
        await this.markThreadAsRead(thread.threadId);

        // Show mobile conversation view
        if (window.innerWidth <= 768) {
            this.showMobileConversation();
        }
    }

    updateActiveThread() {
        // Update thread list UI
        core.$$('.thread-item').forEach(item => {
            item.classList.toggle('active', 
                item.dataset.threadId === this.currentThread?.threadId);
        });

        // Update conversation header
        const header = core.$('#conversationHeader h3');
        if (header && this.currentThread) {
            header.textContent = this.currentThread.subject;
        }
    }

    showConversationArea() {
        const conversationArea = core.$('.message-conversation');
        const noConversation = core.$('.no-conversation');
        const messageInput = core.$('#messageInput');

        if (conversationArea) {
            conversationArea.style.display = 'flex';
        }
        
        if (noConversation) {
            noConversation.style.display = 'none';
        }
        
        if (messageInput) {
            messageInput.style.display = 'block';
        }
    }

    async loadThreadMessages(threadId) {
        try {
            this.currentMessages = await api.getThreadMessages(threadId, 50, 0);
            this.renderMessages();
        } catch (error) {
            console.error('Error loading thread messages:', error);
            // Use mock data for development
            this.currentMessages = this.getMockMessages(threadId);
            this.renderMessages();
        }
    }

    renderMessages() {
        const container = core.$('#conversationMessages');
        if (!container) return;

        container.innerHTML = '';

        if (this.currentMessages.length === 0) {
            container.innerHTML = `
                <div class="no-messages">
                    <p>No messages in this conversation yet.</p>
                    <p>Start the conversation by sending a message below.</p>
                </div>
            `;
            return;
        }

        this.currentMessages.forEach(message => {
            const messageElement = this.createMessageElement(message);
            container.appendChild(messageElement);
        });

        // Scroll to bottom
        this.scrollToBottom();
    }

    createMessageElement(message) {
        const element = core.createElement('div', {
            className: `message-item ${message.type.toLowerCase()}`,
            'data-message-id': message.messageId
        });

        const time = core.formatTime(message.createdAt);

        element.innerHTML = `
            <div class="message-bubble">
                ${message.content}
            </div>
            <div class="message-time">${time}</div>
        `;

        return element;
    }

    async sendMessage() {
        const messageInput = core.$('#messageText');
        if (!messageInput || !this.currentThread) return;

        const content = messageInput.value.trim();
        if (!content) return;

        try {
            // Clear input immediately for better UX
            messageInput.value = '';

            // Send message
            const newMessage = await api.sendMessage(
                this.currentThread.threadId, 
                'Client', 
                content
            );

            // Add to current messages
            this.currentMessages.push(newMessage);
            this.renderMessages();

            // Update thread list
            this.updateThreadLastMessage(this.currentThread.threadId);

        } catch (error) {
            // Restore input value on error
            messageInput.value = content;
            core.handleError(error, 'Send Message');
        }
    }

    updateThreadLastMessage(threadId) {
        const thread = this.messageThreads.find(t => t.threadId === threadId);
        if (thread) {
            thread.lastMessage = new Date().toISOString();
            this.renderThreads();
        }
    }

    async markThreadAsRead(threadId) {
        try {
            await api.markMessagesAsRead(threadId);
            
            // Update UI
            const threadElement = core.$(`[data-thread-id="${threadId}"]`);
            if (threadElement) {
                threadElement.classList.remove('unread');
            }

        } catch (error) {
            console.error('Error marking messages as read:', error);
        }
    }

    showNewThreadModal() {
        // Load projects for selection
        this.loadProjectsForNewThread();
        
        const modal = core.$('#newThreadModal');
        if (modal) {
            modal.classList.add('show');
        }
    }

    async loadProjectsForNewThread() {
        const projectSelect = core.$('#threadProjectSelect');
        if (!projectSelect) return;

        try {
            // This would normally get projects from the API
            const projects = this.getMockProjects();
            
            projectSelect.innerHTML = '<option value="">Select a project</option>';
            projects.forEach(project => {
                const option = core.createElement('option', {
                    value: project.projectId
                }, project.title);
                projectSelect.appendChild(option);
            });

        } catch (error) {
            console.error('Error loading projects:', error);
        }
    }

    async createNewThread() {
        const projectSelect = core.$('#threadProjectSelect');
        const subjectInput = core.$('#threadSubject');
        const messageInput = core.$('#threadMessage');

        if (!projectSelect || !subjectInput || !messageInput) return;

        const projectId = projectSelect.value;
        const subject = subjectInput.value.trim();
        const message = messageInput.value.trim();

        if (!projectId || !subject || !message) {
            core.showToast('Please fill in all fields', 'warning');
            return;
        }

        try {
            // Create thread
            const newThread = await api.createMessageThread(
                projectId, 
                core.currentClientId, 
                subject
            );

            // Send initial message
            await api.sendMessage(newThread.threadId, 'Client', message);

            // Close modal
            this.closeModal();

            // Refresh threads
            await this.loadData();

            // Select new thread
            this.selectThread(newThread);

            core.handleSuccess('Conversation started successfully');

        } catch (error) {
            core.handleError(error, 'Create Conversation');
        }
    }

    closeModal() {
        const modal = core.$('#newThreadModal');
        if (modal) {
            modal.classList.remove('show');
            
            // Reset form
            const form = modal.querySelector('#newThreadForm');
            if (form) {
                form.reset();
            }
        }
    }

    showMobileConversation() {
        const messagesLayout = core.$('.messages-layout');
        if (messagesLayout) {
            messagesLayout.classList.add('show-conversation');
        }
    }

    showMobileThreads() {
        const messagesLayout = core.$('.messages-layout');
        if (messagesLayout) {
            messagesLayout.classList.remove('show-conversation');
        }
    }

    scrollToBottom() {
        const container = core.$('#conversationMessages');
        if (container) {
            container.scrollTop = container.scrollHeight;
        }
    }

    startPolling() {
        this.pollInterval = setInterval(() => {
            if (document.visibilityState === 'visible' && this.currentThread) {
                this.loadThreadMessages(this.currentThread.threadId);
            }
        }, this.pollRate);
    }

    stopPolling() {
        if (this.pollInterval) {
            clearInterval(this.pollInterval);
            this.pollInterval = null;
        }
    }

    // Mock data methods
    getMockMessageThreads() {
        return [
            {
                threadId: 'thread-1',
                projectId: 'proj-123',
                clientId: core.currentClientId,
                subject: 'SEO Content Requirements',
                isActive: true,
                lastMessage: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
                createdAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
            },
            {
                threadId: 'thread-2',
                projectId: 'proj-124',
                clientId: core.currentClientId,
                subject: 'Social Media Content Calendar',
                isActive: true,
                lastMessage: new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString(),
                createdAt: new Date(Date.now() - 48 * 60 * 60 * 1000).toISOString()
            },
            {
                threadId: 'thread-3',
                projectId: 'proj-125',
                clientId: core.currentClientId,
                subject: 'Email Campaign Design Feedback',
                isActive: true,
                lastMessage: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
                createdAt: new Date(Date.now() - 72 * 60 * 60 * 1000).toISOString()
            }
        ];
    }

    getMockMessages(threadId) {
        const messages = {
            'thread-1': [
                {
                    messageId: 'msg-1',
                    threadId: threadId,
                    type: 'System',
                    content: 'Hello! I\'m here to help with your SEO content project. What specific requirements do you have?',
                    isRead: true,
                    createdAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
                },
                {
                    messageId: 'msg-2',
                    threadId: threadId,
                    type: 'Client',
                    content: 'Hi! I need the content to target these keywords: "content marketing", "SEO strategy", and "digital growth". The tone should be professional but approachable.',
                    isRead: true,
                    createdAt: new Date(Date.now() - 23 * 60 * 60 * 1000).toISOString()
                },
                {
                    messageId: 'msg-3',
                    threadId: threadId,
                    type: 'System',
                    content: 'Perfect! I\'ll incorporate those keywords naturally throughout the content. Do you have any specific word count requirements or preferred content structure?',
                    isRead: true,
                    createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString()
                }
            ],
            'thread-2': [
                {
                    messageId: 'msg-4',
                    threadId: threadId,
                    type: 'System',
                    content: 'I\'ve prepared your social media content calendar. Could you review the themes and posting schedule?',
                    isRead: true,
                    createdAt: new Date(Date.now() - 48 * 60 * 60 * 1000).toISOString()
                },
                {
                    messageId: 'msg-5',
                    threadId: threadId,
                    type: 'Client',
                    content: 'Thanks! The themes look great. Can we adjust the posting frequency to 3 times per week instead of daily?',
                    isRead: true,
                    createdAt: new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString()
                }
            ]
        };

        return messages[threadId] || [];
    }

    getMockThreadPreview(subject) {
        const previews = {
            'SEO Content Requirements': 'Hi! I need the content to target these keywords...',
            'Social Media Content Calendar': 'Thanks! The themes look great. Can we adjust...',
            'Email Campaign Design Feedback': 'The design looks fantastic! Just a few minor...'
        };
        return previews[subject] || 'Click to view conversation';
    }

    getMockProjects() {
        return [
            { projectId: 'proj-123', title: 'SEO Content Package' },
            { projectId: 'proj-124', title: 'Social Media Strategy' },
            { projectId: 'proj-125', title: 'Email Marketing Campaign' }
        ];
    }

    destroy() {
        this.stopPolling();
    }
}