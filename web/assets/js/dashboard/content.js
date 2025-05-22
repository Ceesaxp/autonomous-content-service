// Dashboard Content Module

import { core } from './core.js';
import { api } from './api.js';

export class ContentManager {
    constructor() {
        this.contentApprovals = [];
        this.filteredContent = [];
        this.currentFilter = '';
    }

    async init() {
        this.setupEventListeners();
        await this.loadData();
    }

    setupEventListeners() {
        // Filter functionality
        const filterSelect = core.$('#contentFilter');
        if (filterSelect) {
            filterSelect.addEventListener('change', (e) => {
                this.currentFilter = e.target.value;
                this.filterContent();
            });
        }

        // Modal handlers
        this.setupModalHandlers();
    }

    setupModalHandlers() {
        const modal = core.$('#contentApprovalModal');
        if (!modal) return;

        // Close button
        const closeBtn = modal.querySelector('.modal-close');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => {
                this.closeModal();
            });
        }

        // Click outside to close
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.closeModal();
            }
        });
    }

    async loadData() {
        try {
            this.contentApprovals = await api.getContentApprovals(core.currentClientId, 50, 0);
            this.filterContent();
        } catch (error) {
            console.error('Error loading content approvals:', error);
            // Use mock data for development
            this.contentApprovals = api.getMockContentApprovals();
            this.filterContent();
        }
    }

    filterContent() {
        let filtered = [...this.contentApprovals];

        // Apply status filter
        if (this.currentFilter) {
            filtered = filtered.filter(content => content.status === this.currentFilter);
        }

        this.filteredContent = filtered;
        this.renderContent();
    }

    renderContent() {
        const container = core.$('#contentApprovalsList');
        if (!container) return;

        container.innerHTML = '';

        if (this.filteredContent.length === 0) {
            container.innerHTML = this.renderEmptyState();
            return;
        }

        this.filteredContent.forEach(content => {
            const contentElement = this.createContentCard(content);
            container.appendChild(contentElement);
        });
    }

    createContentCard(approval) {
        const card = core.createElement('div', {
            className: 'content-approval-card',
            'data-approval-id': approval.approvalId
        });

        const statusClass = approval.status.toLowerCase();
        const timeAgo = core.getRelativeTime(approval.createdAt);

        // Mock content details for development
        const content = approval.content || {
            title: 'Content Title',
            type: 'blog_post',
            wordCount: 1200
        };

        card.innerHTML = `
            <div class="content-header">
                <div class="content-info">
                    <h3 class="content-title">${content.title}</h3>
                    <p class="content-project">Project: ${this.getProjectName(approval.projectId)}</p>
                    <div class="content-meta">
                        <span class="content-type">${this.formatContentType(content.type)}</span>
                        <span class="content-length">${content.wordCount} words</span>
                        <span class="content-time">${timeAgo}</span>
                    </div>
                </div>
                <div class="content-status-wrapper">
                    <span class="content-status ${statusClass}">${approval.status}</span>
                </div>
            </div>

            ${approval.feedback ? `
                <div class="content-feedback">
                    <strong>Feedback:</strong> ${approval.feedback}
                </div>
            ` : ''}

            <div class="content-actions">
                <button class="btn btn-secondary" data-action="preview">Preview</button>
                ${approval.status === 'Pending' ? `
                    <button class="btn btn-success" data-action="approve">Approve</button>
                    <button class="btn btn-danger" data-action="reject">Reject</button>
                    <button class="btn btn-primary" data-action="revision">Request Revision</button>
                ` : approval.status === 'Revision' ? `
                    <button class="btn btn-success" data-action="approve">Approve</button>
                    <button class="btn btn-primary" data-action="revision">More Changes</button>
                ` : ''}
            </div>
        `;

        // Add click handlers
        this.setupCardActions(card, approval);

        return card;
    }

    setupCardActions(card, approval) {
        // Preview action
        const previewBtn = card.querySelector('[data-action="preview"]');
        if (previewBtn) {
            previewBtn.addEventListener('click', () => {
                this.showContentPreview(approval);
            });
        }

        // Approve action
        const approveBtn = card.querySelector('[data-action="approve"]');
        if (approveBtn) {
            approveBtn.addEventListener('click', () => {
                this.showApprovalDialog(approval, 'approve');
            });
        }

        // Reject action
        const rejectBtn = card.querySelector('[data-action="reject"]');
        if (rejectBtn) {
            rejectBtn.addEventListener('click', () => {
                this.showApprovalDialog(approval, 'reject');
            });
        }

        // Revision action
        const revisionBtn = card.querySelector('[data-action="revision"]');
        if (revisionBtn) {
            revisionBtn.addEventListener('click', () => {
                this.showApprovalDialog(approval, 'revision');
            });
        }
    }

    showContentPreview(approval) {
        const content = approval.content || { title: 'Content Title', body: 'Content preview not available.' };
        
        this.renderContentModal({
            title: 'Content Preview',
            content: `
                <div class="content-preview">
                    <h3>${content.title}</h3>
                    <div class="content-meta">
                        <span>Type: ${this.formatContentType(content.type)}</span>
                        <span>Word Count: ${content.wordCount || 'N/A'}</span>
                        <span>Created: ${core.formatDate(approval.createdAt)}</span>
                    </div>
                    <div class="content-body">
                        ${content.body || this.getMockContentBody(content.type)}
                    </div>
                </div>
            `,
            actions: `
                <button class="btn btn-secondary" data-modal-close>Close</button>
                ${approval.status === 'Pending' ? `
                    <button class="btn btn-success" data-action="approve">Approve</button>
                    <button class="btn btn-primary" data-action="revision">Request Revision</button>
                ` : ''}
            `,
            approval: approval
        });

        this.openModal();
    }

    showApprovalDialog(approval, action) {
        const titles = {
            approve: 'Approve Content',
            reject: 'Reject Content',
            revision: 'Request Revision'
        };

        const placeholders = {
            approve: 'Optional: Add approval comments...',
            reject: 'Please explain why this content is being rejected...',
            revision: 'Please describe what changes are needed...'
        };

        const buttonTexts = {
            approve: 'Approve',
            reject: 'Reject',
            revision: 'Request Revision'
        };

        const required = action !== 'approve';

        this.renderContentModal({
            title: titles[action],
            content: `
                <div class="approval-dialog">
                    <h4>${approval.content?.title || 'Content Title'}</h4>
                    <div class="form-group">
                        <label for="feedbackText">Feedback</label>
                        <textarea id="feedbackText" 
                                placeholder="${placeholders[action]}"
                                ${required ? 'required' : ''}
                                rows="4"></textarea>
                    </div>
                </div>
            `,
            actions: `
                <button class="btn btn-secondary" data-modal-close>Cancel</button>
                <button class="btn ${action === 'approve' ? 'btn-success' : action === 'reject' ? 'btn-danger' : 'btn-primary'}" 
                        data-action="submit">${buttonTexts[action]}</button>
            `,
            onSubmit: (feedback) => this.handleApprovalAction(approval, action, feedback)
        });

        this.openModal();
    }

    async handleApprovalAction(approval, action, feedback) {
        try {
            switch (action) {
                case 'approve':
                    await api.approveContent(approval.approvalId, feedback);
                    core.handleSuccess('Content approved successfully');
                    break;
                case 'reject':
                    if (!feedback.trim()) {
                        core.showToast('Please provide feedback for rejection', 'warning');
                        return;
                    }
                    await api.rejectContent(approval.approvalId, feedback);
                    core.handleSuccess('Content rejected');
                    break;
                case 'revision':
                    if (!feedback.trim()) {
                        core.showToast('Please describe the required changes', 'warning');
                        return;
                    }
                    await api.requestContentRevision(approval.approvalId, feedback);
                    core.handleSuccess('Revision requested');
                    break;
            }

            this.closeModal();
            await this.loadData(); // Refresh data

        } catch (error) {
            core.handleError(error, 'Content Approval Action');
        }
    }

    renderContentModal({ title, content, actions, onSubmit, approval }) {
        const modal = core.$('#contentApprovalModal');
        if (!modal) return;

        const modalTitle = modal.querySelector('.modal-header h2');
        const modalBody = modal.querySelector('.modal-body');

        if (modalTitle) modalTitle.textContent = title;
        if (modalBody) {
            modalBody.innerHTML = `
                ${content}
                ${actions ? `
                    <div class="modal-actions">
                        ${actions}
                    </div>
                ` : ''}
            `;

            // Setup action handlers
            const closeButtons = modalBody.querySelectorAll('[data-modal-close]');
            closeButtons.forEach(btn => {
                btn.addEventListener('click', () => this.closeModal());
            });

            const submitButton = modalBody.querySelector('[data-action="submit"]');
            if (submitButton && onSubmit) {
                submitButton.addEventListener('click', () => {
                    const feedbackText = modalBody.querySelector('#feedbackText');
                    const feedback = feedbackText ? feedbackText.value : '';
                    onSubmit(feedback);
                });
            }

            // Setup approval actions in preview
            const approveBtn = modalBody.querySelector('[data-action="approve"]');
            const revisionBtn = modalBody.querySelector('[data-action="revision"]');
            
            if (approveBtn && approval) {
                approveBtn.addEventListener('click', () => {
                    this.showApprovalDialog(approval, 'approve');
                });
            }
            
            if (revisionBtn && approval) {
                revisionBtn.addEventListener('click', () => {
                    this.showApprovalDialog(approval, 'revision');
                });
            }
        }
    }

    renderEmptyState() {
        const message = this.currentFilter ? 
            'No content matches your filter criteria.' : 
            'No content approvals found.';

        return `
            <div class="empty-state">
                <div class="empty-icon">📝</div>
                <h3>No Content Found</h3>
                <p>${message}</p>
                ${this.currentFilter ? 
                    '<button class="btn btn-secondary" onclick="this.clearFilters()">Clear Filters</button>' : 
                    ''
                }
            </div>
        `;
    }

    openModal() {
        const modal = core.$('#contentApprovalModal');
        if (modal) {
            modal.classList.add('show');
        }
    }

    closeModal() {
        const modal = core.$('#contentApprovalModal');
        if (modal) {
            modal.classList.remove('show');
        }
    }

    clearFilters() {
        this.currentFilter = '';
        const filterSelect = core.$('#contentFilter');
        if (filterSelect) filterSelect.value = '';
        this.filterContent();
    }

    // Helper methods
    getProjectName(projectId) {
        // This would typically come from the projects data
        const projectNames = {
            'proj-123': 'SEO Content Package',
            'proj-124': 'Social Media Strategy',
            'proj-125': 'Email Marketing Campaign'
        };
        return projectNames[projectId] || 'Unknown Project';
    }

    formatContentType(type) {
        const typeMap = {
            'blog_post': 'Blog Post',
            'landing_page': 'Landing Page',
            'email': 'Email',
            'social_media': 'Social Media',
            'copy': 'Copy',
            'template': 'Template',
            'guide': 'Guide'
        };
        return typeMap[type] || type;
    }

    getMockContentBody(type) {
        const mockBodies = {
            'blog_post': `
                <h2>Introduction</h2>
                <p>This is a sample blog post content that demonstrates the preview functionality. The content would typically be much longer and include proper formatting, images, and links.</p>
                
                <h3>Key Points</h3>
                <ul>
                    <li>First important point about the topic</li>
                    <li>Second key insight for readers</li>
                    <li>Third actionable takeaway</li>
                </ul>
                
                <p>The conclusion wraps up the main ideas and provides a clear call to action for readers.</p>
            `,
            'email': `
                <p><strong>Subject:</strong> Your Weekly Content Marketing Tips</p>
                
                <p>Hi [First Name],</p>
                
                <p>Hope you're having a great week! Here are this week's top content marketing insights:</p>
                
                <ol>
                    <li>How to optimize your content for voice search</li>
                    <li>The importance of content repurposing</li>
                    <li>Measuring content ROI effectively</li>
                </ol>
                
                <p>Best regards,<br>Your Content Team</p>
            `,
            'default': `
                <p>This is a sample content preview. The actual content would appear here with proper formatting and structure.</p>
            `
        };
        
        return mockBodies[type] || mockBodies['default'];
    }
}