// Dashboard Projects Module

import { core } from './core.js';
import { api } from './api.js';

export class ProjectManager {
    constructor() {
        this.projects = [];
        this.filteredProjects = [];
        this.currentFilter = '';
        this.searchTerm = '';
    }

    async init() {
        this.setupEventListeners();
        await this.loadData();
    }

    setupEventListeners() {
        // Search functionality
        const searchInput = core.$('#projectsSearch');
        if (searchInput) {
            searchInput.addEventListener('input', core.debounce((e) => {
                this.searchTerm = e.target.value;
                this.filterProjects();
            }, 300));
        }

        // Filter functionality
        const filterSelect = core.$('#projectsFilter');
        if (filterSelect) {
            filterSelect.addEventListener('change', (e) => {
                this.currentFilter = e.target.value;
                this.filterProjects();
            });
        }

        // Modal close handlers
        this.setupModalHandlers();

        // Listen for navigation events
        core.on('show-project-details', (e) => {
            this.showProjectDetails(e.detail.projectId);
        });
    }

    setupModalHandlers() {
        const modal = core.$('#projectDetailsModal');
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
            this.projects = await api.getProjectsOverview(core.currentClientId);
            this.filterProjects();
        } catch (error) {
            console.error('Error loading projects:', error);
            // Use mock data for development
            this.projects = api.getMockProjectsOverview();
            this.filterProjects();
        }
    }

    filterProjects() {
        let filtered = [...this.projects];

        // Apply status filter
        if (this.currentFilter) {
            filtered = filtered.filter(project => project.status === this.currentFilter);
        }

        // Apply search filter
        if (this.searchTerm) {
            const search = this.searchTerm.toLowerCase();
            filtered = filtered.filter(project => 
                project.title.toLowerCase().includes(search) ||
                project.status.toLowerCase().includes(search) ||
                project.priority.toLowerCase().includes(search)
            );
        }

        this.filteredProjects = filtered;
        this.renderProjects();
    }

    renderProjects() {
        const container = core.$('#projectsList');
        if (!container) return;

        container.innerHTML = '';

        if (this.filteredProjects.length === 0) {
            container.innerHTML = this.renderEmptyState();
            return;
        }

        this.filteredProjects.forEach(project => {
            const projectElement = this.createProjectCard(project);
            container.appendChild(projectElement);
        });
    }

    createProjectCard(project) {
        const card = core.createElement('div', {
            className: 'project-card',
            'data-project-id': project.projectId
        });

        const daysText = project.daysRemaining > 0 ? 
            `${project.daysRemaining} days remaining` : 
            project.daysRemaining === 0 ? 'Due today' : 'Overdue';

        const statusClass = project.status.toLowerCase().replace(/\s+/g, '-');
        const priorityClass = project.priority.toLowerCase();

        card.innerHTML = `
            <div class="project-header">
                <div class="project-info">
                    <h3 class="project-title">${project.title}</h3>
                    <div class="project-meta">
                        <span class="project-status ${statusClass}">${project.status}</span>
                        <span class="project-priority ${priorityClass}">${project.priority}</span>
                    </div>
                </div>
                <div class="project-actions">
                    <button class="btn btn-secondary" onclick="event.stopPropagation()" data-action="details">
                        View Details
                    </button>
                </div>
            </div>
            
            <div class="project-progress">
                <div class="progress-label">
                    <span>Progress</span>
                    <span>${project.progress.toFixed(0)}%</span>
                </div>
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${project.progress}%"></div>
                </div>
            </div>
            
            <div class="project-footer">
                <div class="project-stats">
                    <span title="Content pieces">${project.contentCount} content</span>
                    <span title="Pending approvals">${project.pendingApprovals} pending</span>
                    <span title="Budget">${core.formatCurrency(project.budget.amount)}</span>
                </div>
                <div class="project-deadline">
                    <span class="${project.daysRemaining < 0 ? 'overdue' : project.daysRemaining <= 3 ? 'urgent' : ''}">${daysText}</span>
                </div>
            </div>
        `;

        // Add click handlers
        card.addEventListener('click', () => {
            this.showProjectDetails(project.projectId);
        });

        const detailsBtn = card.querySelector('[data-action="details"]');
        if (detailsBtn) {
            detailsBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.showProjectDetails(project.projectId);
            });
        }

        return card;
    }

    renderEmptyState() {
        const message = this.searchTerm || this.currentFilter ? 
            'No projects match your search criteria.' : 
            'No projects found.';

        return `
            <div class="empty-state">
                <div class="empty-icon">📋</div>
                <h3>No Projects Found</h3>
                <p>${message}</p>
                ${this.searchTerm || this.currentFilter ? 
                    '<button class="btn btn-secondary" onclick="this.clearFilters()">Clear Filters</button>' : 
                    ''
                }
            </div>
        `;
    }

    async showProjectDetails(projectId) {
        try {
            const projectDetails = await this.getProjectDetails(projectId);
            this.renderProjectDetailsModal(projectDetails);
            this.openModal();
        } catch (error) {
            core.handleError(error, 'Load Project Details');
        }
    }

    async getProjectDetails(projectId) {
        try {
            return await api.getProjectDetails(projectId);
        } catch (error) {
            // Return mock data for development
            return this.getMockProjectDetails(projectId);
        }
    }

    renderProjectDetailsModal(details) {
        const modal = core.$('#projectDetailsModal');
        if (!modal) return;

        const modalBody = modal.querySelector('.modal-body');
        if (!modalBody) return;

        const project = details.project;
        const milestones = details.milestones || [];
        const approvals = details.contentApprovals || [];
        const analytics = details.analytics;

        modalBody.innerHTML = `
            <div class="project-details-content">
                <!-- Project Overview -->
                <div class="project-overview">
                    <div class="overview-header">
                        <h3>${project.title}</h3>
                        <div class="project-badges">
                            <span class="badge status-${project.status.toLowerCase()}">${project.status}</span>
                            <span class="badge priority-${project.priority.toLowerCase()}">${project.priority}</span>
                        </div>
                    </div>
                    
                    <div class="overview-stats">
                        <div class="stat">
                            <label>Budget</label>
                            <value>${core.formatCurrency(project.budget.amount)}</value>
                        </div>
                        <div class="stat">
                            <label>Deadline</label>
                            <value>${core.formatDate(project.deadline)}</value>
                        </div>
                        <div class="stat">
                            <label>Content Pieces</label>
                            <value>${project.contents?.length || 0}</value>
                        </div>
                        <div class="stat">
                            <label>Progress</label>
                            <value>${analytics?.progressPercent?.toFixed(0) || 0}%</value>
                        </div>
                    </div>
                </div>

                <!-- Project Description -->
                <div class="project-description">
                    <h4>Description</h4>
                    <p>${project.description}</p>
                </div>

                <!-- Project Requirements -->
                ${project.requirements && project.requirements.length > 0 ? `
                    <div class="project-requirements">
                        <h4>Requirements</h4>
                        <ul>
                            ${project.requirements.map(req => `<li>${req}</li>`).join('')}
                        </ul>
                    </div>
                ` : ''}

                <!-- Milestones -->
                ${milestones.length > 0 ? `
                    <div class="project-milestones">
                        <h4>Milestones</h4>
                        <div class="milestones-list">
                            ${milestones.map(milestone => this.renderMilestone(milestone)).join('')}
                        </div>
                    </div>
                ` : ''}

                <!-- Content Approvals -->
                ${approvals.length > 0 ? `
                    <div class="project-approvals">
                        <h4>Content Approvals</h4>
                        <div class="approvals-list">
                            ${approvals.map(approval => this.renderApproval(approval)).join('')}
                        </div>
                    </div>
                ` : ''}

                <!-- Actions -->
                <div class="project-actions">
                    <button class="btn btn-primary" data-action="update-status">Update Status</button>
                    <button class="btn btn-secondary" data-action="view-content">View Content</button>
                    <button class="btn btn-secondary" data-action="send-message">Send Message</button>
                </div>
            </div>
        `;

        // Add action handlers
        this.setupProjectDetailsActions(project);
    }

    renderMilestone(milestone) {
        const statusIcon = milestone.isCompleted ? '✅' : '⏳';
        const statusClass = milestone.isCompleted ? 'completed' : 'pending';
        
        return `
            <div class="milestone-item ${statusClass}">
                <div class="milestone-status">${statusIcon}</div>
                <div class="milestone-content">
                    <div class="milestone-title">${milestone.title}</div>
                    <div class="milestone-description">${milestone.description}</div>
                    <div class="milestone-date">Due: ${core.formatDate(milestone.dueDate)}</div>
                </div>
            </div>
        `;
    }

    renderApproval(approval) {
        const statusColors = {
            'Pending': 'warning',
            'Approved': 'success',
            'Rejected': 'danger',
            'Revision': 'info'
        };
        
        return `
            <div class="approval-item">
                <div class="approval-content">
                    <div class="approval-title">Content Approval</div>
                    <div class="approval-status status-${statusColors[approval.status]}">${approval.status}</div>
                </div>
                <div class="approval-date">${core.formatDate(approval.createdAt)}</div>
            </div>
        `;
    }

    setupProjectDetailsActions(project) {
        const modal = core.$('#projectDetailsModal');
        if (!modal) return;

        // Update status action
        const updateStatusBtn = modal.querySelector('[data-action="update-status"]');
        if (updateStatusBtn) {
            updateStatusBtn.addEventListener('click', () => {
                this.showUpdateStatusDialog(project);
            });
        }

        // View content action
        const viewContentBtn = modal.querySelector('[data-action="view-content"]');
        if (viewContentBtn) {
            viewContentBtn.addEventListener('click', () => {
                this.closeModal();
                core.emit('navigate', { view: 'content' });
            });
        }

        // Send message action
        const sendMessageBtn = modal.querySelector('[data-action="send-message"]');
        if (sendMessageBtn) {
            sendMessageBtn.addEventListener('click', () => {
                this.closeModal();
                core.emit('navigate', { view: 'messages' });
                // Could also trigger a new message modal
            });
        }
    }

    showUpdateStatusDialog(project) {
        const statuses = ['Draft', 'Planning', 'InProgress', 'Review', 'Completed', 'Cancelled'];
        const currentStatus = project.status;
        
        const dialog = core.createElement('div', { className: 'status-dialog' }, `
            <h4>Update Project Status</h4>
            <select id="statusSelect" class="status-select">
                ${statuses.map(status => 
                    `<option value="${status}" ${status === currentStatus ? 'selected' : ''}>${status}</option>`
                ).join('')}
            </select>
            <div class="dialog-actions">
                <button class="btn btn-secondary" data-action="cancel">Cancel</button>
                <button class="btn btn-primary" data-action="update">Update</button>
            </div>
        `);

        const modalBody = core.$('#projectDetailsModal .modal-body');
        if (modalBody) {
            modalBody.appendChild(dialog);
        }

        // Add handlers
        const cancelBtn = dialog.querySelector('[data-action="cancel"]');
        const updateBtn = dialog.querySelector('[data-action="update"]');
        
        if (cancelBtn) {
            cancelBtn.addEventListener('click', () => {
                dialog.remove();
            });
        }

        if (updateBtn) {
            updateBtn.addEventListener('click', async () => {
                const newStatus = dialog.querySelector('#statusSelect').value;
                await this.updateProjectStatus(project.projectId, newStatus);
                dialog.remove();
                this.closeModal();
            });
        }
    }

    async updateProjectStatus(projectId, status) {
        try {
            await api.updateProjectStatus(projectId, status);
            core.handleSuccess('Project status updated successfully');
            
            // Refresh data
            await this.loadData();
            
        } catch (error) {
            core.handleError(error, 'Update Project Status');
        }
    }

    openModal() {
        const modal = core.$('#projectDetailsModal');
        if (modal) {
            modal.classList.add('show');
        }
    }

    closeModal() {
        const modal = core.$('#projectDetailsModal');
        if (modal) {
            modal.classList.remove('show');
        }
    }

    clearFilters() {
        this.currentFilter = '';
        this.searchTerm = '';
        
        const searchInput = core.$('#projectsSearch');
        const filterSelect = core.$('#projectsFilter');
        
        if (searchInput) searchInput.value = '';
        if (filterSelect) filterSelect.value = '';
        
        this.filterProjects();
    }

    getMockProjectDetails(projectId) {
        const project = this.projects.find(p => p.projectId === projectId);
        if (!project) return null;

        return {
            project: {
                ...project,
                description: 'Comprehensive SEO content package including blog posts, landing pages, and meta descriptions to improve search engine visibility and drive organic traffic.',
                requirements: [
                    'Target keywords: SEO, content marketing, digital strategy',
                    'Tone: Professional yet approachable',
                    'Word count: 1500-2000 words per piece',
                    'Include call-to-action buttons',
                    'SEO optimized with meta descriptions'
                ],
                contents: [
                    { id: 'content-1', title: 'SEO Best Practices 2024', type: 'blog_post', status: 'completed' },
                    { id: 'content-2', title: 'Technical SEO Guide', type: 'blog_post', status: 'in_progress' },
                    { id: 'content-3', title: 'Landing Page Copy', type: 'copy', status: 'pending' }
                ]
            },
            milestones: [
                {
                    milestoneId: 'milestone-1',
                    projectId: projectId,
                    title: 'Content Strategy Approval',
                    description: 'Review and approve the content strategy and outline',
                    dueDate: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
                    isCompleted: true,
                    completedAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString()
                },
                {
                    milestoneId: 'milestone-2',
                    projectId: projectId,
                    title: 'First Draft Delivery',
                    description: 'Deliver first drafts of all content pieces',
                    dueDate: new Date(Date.now() + 2 * 24 * 60 * 60 * 1000).toISOString(),
                    isCompleted: false
                },
                {
                    milestoneId: 'milestone-3',
                    projectId: projectId,
                    title: 'Final Delivery',
                    description: 'Deliver final versions after revisions',
                    dueDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
                    isCompleted: false
                }
            ],
            contentApprovals: [
                {
                    approvalId: 'approval-1',
                    contentId: 'content-1',
                    status: 'Approved',
                    createdAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
                },
                {
                    approvalId: 'approval-2',
                    contentId: 'content-2',
                    status: 'Pending',
                    createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString()
                }
            ],
            analytics: {
                projectId: projectId,
                progressPercent: project.progress,
                totalTasks: 8,
                completedTasks: 6,
                pendingTasks: 2,
                timeSpent: '24 hours',
                estimatedTime: '32 hours',
                daysToDeadline: project.daysRemaining
            }
        };
    }
}