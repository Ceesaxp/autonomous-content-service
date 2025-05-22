// Dashboard Analytics Module

import { core } from './core.js';
import { api } from './api.js';

export class AnalyticsManager {
    constructor() {
        this.analyticsData = null;
        this.currentDateRange = {
            from: this.getDefaultFromDate(),
            to: this.getDefaultToDate()
        };
        this.charts = {};
    }

    async init() {
        this.setupEventListeners();
        this.initializeDateInputs();
        await this.loadData();
    }

    setupEventListeners() {
        // Update analytics button
        const updateBtn = core.$('#updateAnalyticsBtn');
        if (updateBtn) {
            updateBtn.addEventListener('click', () => {
                this.updateDateRange();
            });
        }

        // Generate report button
        const generateReportBtn = core.$('#generateReportBtn');
        if (generateReportBtn) {
            generateReportBtn.addEventListener('click', () => {
                this.generateReport();
            });
        }

        // Date input changes
        const fromDate = core.$('#analyticsFromDate');
        const toDate = core.$('#analyticsToDate');

        if (fromDate) {
            fromDate.addEventListener('change', () => {
                this.validateDateRange();
            });
        }

        if (toDate) {
            toDate.addEventListener('change', () => {
                this.validateDateRange();
            });
        }
    }

    initializeDateInputs() {
        const fromDateInput = core.$('#analyticsFromDate');
        const toDateInput = core.$('#analyticsToDate');

        if (fromDateInput) {
            fromDateInput.value = this.formatDateForInput(this.currentDateRange.from);
        }

        if (toDateInput) {
            toDateInput.value = this.formatDateForInput(this.currentDateRange.to);
        }
    }

    async loadData() {
        try {
            this.analyticsData = await api.getClientAnalytics(
                core.currentClientId,
                this.formatDateForAPI(this.currentDateRange.from),
                this.formatDateForAPI(this.currentDateRange.to)
            );
            
            this.renderAnalytics();
            
        } catch (error) {
            console.error('Error loading analytics:', error);
            // Use mock data for development
            this.analyticsData = this.getMockAnalytics();
            this.renderAnalytics();
        }
    }

    renderAnalytics() {
        this.renderPerformanceMetrics();
        this.renderMonthlyActivity();
        this.renderCharts();
    }

    renderPerformanceMetrics() {
        const container = core.$('#performanceMetricsList');
        if (!container || !this.analyticsData) return;

        const metrics = this.analyticsData.performanceMetrics;
        if (!metrics) return;

        container.innerHTML = `
            <div class="metric-item">
                <span class="metric-label">On-Time Delivery Rate</span>
                <span class="metric-value">${metrics.onTimeDeliveryRate}%</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Client Satisfaction</span>
                <span class="metric-value">${metrics.clientSatisfactionAvg}/5.0</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Revision Request Rate</span>
                <span class="metric-value">${metrics.revisionRequestRate}%</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Average Response Time</span>
                <span class="metric-value">${metrics.averageResponseTime}</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Total Projects</span>
                <span class="metric-value">${this.analyticsData.totalProjects}</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Active Projects</span>
                <span class="metric-value">${this.analyticsData.activeProjects}</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Content Delivered</span>
                <span class="metric-value">${this.analyticsData.totalContentPieces}</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Total Investment</span>
                <span class="metric-value">${core.formatCurrency(this.analyticsData.totalSpent.amount)}</span>
            </div>
        `;
    }

    renderMonthlyActivity() {
        const container = core.$('#monthlyActivityList');
        if (!container || !this.analyticsData) return;

        const monthlyData = this.analyticsData.monthlyActivity || [];

        if (monthlyData.length === 0) {
            container.innerHTML = '<p class="no-data">No monthly activity data available</p>';
            return;
        }

        container.innerHTML = `
            <div class="activity-chart-container">
                <div class="chart-header">
                    <span>Month</span>
                    <span>Projects</span>
                    <span>Content</span>
                    <span>Spent</span>
                </div>
                ${monthlyData.map(month => `
                    <div class="activity-row">
                        <span class="month-name">${month.month}</span>
                        <span class="projects-count">${month.projectsStarted}/${month.projectsCompleted}</span>
                        <span class="content-count">${month.contentDelivered}</span>
                        <span class="amount-spent">${core.formatCurrency(month.amountSpent.amount)}</span>
                    </div>
                `).join('')}
            </div>
        `;
    }

    renderCharts() {
        // Note: In a real implementation, you would use a charting library like Chart.js
        // For this demo, we'll create simple visual representations
        this.renderProjectProgressChart();
        this.renderContentDeliveryChart();
    }

    renderProjectProgressChart() {
        const canvas = core.$('#projectProgressChart');
        if (!canvas || !this.analyticsData) return;

        // Get the container and create a simple chart representation
        const container = canvas.parentElement;
        if (!container) return;

        // Create a simple progress visualization
        const chartData = this.getProjectProgressData();
        const chartHTML = this.createSimpleBarChart(chartData, 'Projects by Status');

        // Replace canvas with HTML chart for demo
        canvas.style.display = 'none';
        
        let existingChart = container.querySelector('.simple-chart');
        if (existingChart) {
            existingChart.remove();
        }

        const chartElement = core.createElement('div', { className: 'simple-chart' }, chartHTML);
        container.appendChild(chartElement);
    }

    renderContentDeliveryChart() {
        const canvas = core.$('#contentDeliveryChart');
        if (!canvas || !this.analyticsData) return;

        const container = canvas.parentElement;
        if (!container) return;

        const chartData = this.getContentDeliveryData();
        const chartHTML = this.createSimplePieChart(chartData, 'Content by Type');

        canvas.style.display = 'none';
        
        let existingChart = container.querySelector('.simple-chart');
        if (existingChart) {
            existingChart.remove();
        }

        const chartElement = core.createElement('div', { className: 'simple-chart' }, chartHTML);
        container.appendChild(chartElement);
    }

    getProjectProgressData() {
        if (!this.analyticsData) return [];

        const statusData = this.analyticsData.projectsByStatus || {};
        return Object.entries(statusData).map(([status, count]) => ({
            label: status,
            value: count,
            color: this.getStatusColor(status)
        }));
    }

    getContentDeliveryData() {
        if (!this.analyticsData) return [];

        const contentData = this.analyticsData.contentByType || {};
        return Object.entries(contentData).map(([type, count]) => ({
            label: this.formatContentType(type),
            value: count,
            color: this.getContentTypeColor(type)
        }));
    }

    createSimpleBarChart(data, title) {
        if (!data || data.length === 0) {
            return '<p class="no-chart-data">No data available for chart</p>';
        }

        const maxValue = Math.max(...data.map(d => d.value));
        
        return `
            <div class="chart-title">${title}</div>
            <div class="bar-chart">
                ${data.map(item => `
                    <div class="bar-item">
                        <div class="bar-label">${item.label}</div>
                        <div class="bar-container">
                            <div class="bar-fill" 
                                 style="width: ${(item.value / maxValue) * 100}%; background-color: ${item.color};">
                            </div>
                            <span class="bar-value">${item.value}</span>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    createSimplePieChart(data, title) {
        if (!data || data.length === 0) {
            return '<p class="no-chart-data">No data available for chart</p>';
        }

        const total = data.reduce((sum, item) => sum + item.value, 0);
        
        return `
            <div class="chart-title">${title}</div>
            <div class="pie-chart-container">
                <div class="pie-chart-legend">
                    ${data.map(item => `
                        <div class="legend-item">
                            <div class="legend-color" style="background-color: ${item.color};"></div>
                            <span class="legend-label">${item.label}</span>
                            <span class="legend-value">${item.value} (${((item.value / total) * 100).toFixed(1)}%)</span>
                        </div>
                    `).join('')}
                </div>
                <div class="pie-chart-visual">
                    ${this.createPieSlices(data, total)}
                </div>
            </div>
        `;
    }

    createPieSlices(data, total) {
        // Simple pie chart representation using CSS
        let currentAngle = 0;
        
        return data.map(item => {
            const percentage = (item.value / total) * 100;
            const angle = (percentage / 100) * 360;
            const slice = `
                <div class="pie-slice" 
                     style="
                        background-color: ${item.color};
                        transform: rotate(${currentAngle}deg);
                        --slice-angle: ${angle}deg;
                     "
                     title="${item.label}: ${item.value}">
                </div>
            `;
            currentAngle += angle;
            return slice;
        }).join('');
    }

    updateDateRange() {
        const fromDateInput = core.$('#analyticsFromDate');
        const toDateInput = core.$('#analyticsToDate');

        if (!fromDateInput || !toDateInput) return;

        const fromDate = new Date(fromDateInput.value);
        const toDate = new Date(toDateInput.value);

        if (!this.isValidDateRange(fromDate, toDate)) {
            core.showToast('Please select a valid date range', 'warning');
            return;
        }

        this.currentDateRange.from = fromDate;
        this.currentDateRange.to = toDate;

        this.loadData();
    }

    validateDateRange() {
        const fromDateInput = core.$('#analyticsFromDate');
        const toDateInput = core.$('#analyticsToDate');
        const updateBtn = core.$('#updateAnalyticsBtn');

        if (!fromDateInput || !toDateInput || !updateBtn) return;

        const fromDate = new Date(fromDateInput.value);
        const toDate = new Date(toDateInput.value);

        const isValid = this.isValidDateRange(fromDate, toDate);
        updateBtn.disabled = !isValid;

        if (!isValid && fromDateInput.value && toDateInput.value) {
            core.showToast('End date must be after start date', 'warning');
        }
    }

    isValidDateRange(fromDate, toDate) {
        return fromDate <= toDate && fromDate <= new Date();
    }

    async generateReport() {
        const reportTypeSelect = core.$('#reportType');
        if (!reportTypeSelect) return;

        const reportType = reportTypeSelect.value;

        try {
            const report = await api.generateReport(core.currentClientId, reportType);
            this.showReportModal(report);
            
        } catch (error) {
            console.error('Error generating report:', error);
            // Show mock report for development
            const mockReport = this.getMockReport(reportType);
            this.showReportModal(mockReport);
        }
    }

    showReportModal(report) {
        // Create a simple report modal
        const modalHTML = `
            <div class="report-modal">
                <div class="report-header">
                    <h3>${report.title}</h3>
                    <p class="report-summary">${report.summary}</p>
                    <p class="report-date">Generated: ${core.formatDateTime(report.generatedAt)}</p>
                </div>
                
                <div class="report-content">
                    ${report.insights && report.insights.length > 0 ? `
                        <div class="report-section">
                            <h4>Key Insights</h4>
                            <ul>
                                ${report.insights.map(insight => `<li>${insight}</li>`).join('')}
                            </ul>
                        </div>
                    ` : ''}
                    
                    ${report.recommendations && report.recommendations.length > 0 ? `
                        <div class="report-section">
                            <h4>Recommendations</h4>
                            <ul>
                                ${report.recommendations.map(rec => `<li>${rec}</li>`).join('')}
                            </ul>
                        </div>
                    ` : ''}
                </div>
                
                <div class="report-actions">
                    <button class="btn btn-secondary" onclick="this.closeReportModal()">Close</button>
                    <button class="btn btn-primary" onclick="this.downloadReport('${report.reportId}')">Download PDF</button>
                </div>
            </div>
        `;

        // Show in a toast for demo purposes
        core.showToast('Report generated successfully! Full report modal would display here.', 'success', 5000);
    }

    // Helper methods
    getDefaultFromDate() {
        const date = new Date();
        date.setMonth(date.getMonth() - 6); // 6 months ago
        return date;
    }

    getDefaultToDate() {
        return new Date();
    }

    formatDateForInput(date) {
        return date.toISOString().split('T')[0];
    }

    formatDateForAPI(date) {
        return date.toISOString().split('T')[0];
    }

    getStatusColor(status) {
        const colors = {
            'InProgress': '#2563eb',
            'Completed': '#10b981',
            'Review': '#f59e0b',
            'Planning': '#8b5cf6',
            'Draft': '#6b7280',
            'Cancelled': '#ef4444'
        };
        return colors[status] || '#6b7280';
    }

    getContentTypeColor(type) {
        const colors = {
            'blog_post': '#3b82f6',
            'email': '#10b981',
            'social_media': '#f59e0b',
            'landing_page': '#8b5cf6',
            'copy': '#ef4444',
            'guide': '#06b6d4'
        };
        return colors[type] || '#6b7280';
    }

    formatContentType(type) {
        const typeMap = {
            'blog_post': 'Blog Posts',
            'email': 'Emails',
            'social_media': 'Social Media',
            'landing_page': 'Landing Pages',
            'copy': 'Copy',
            'guide': 'Guides'
        };
        return typeMap[type] || type;
    }

    // Mock data methods
    getMockAnalytics() {
        return {
            clientId: core.currentClientId,
            totalProjects: 15,
            activeProjects: 5,
            completedProjects: 10,
            totalContentPieces: 45,
            totalSpent: { amount: 12750.00, currency: 'USD' },
            projectsByStatus: {
                'InProgress': 5,
                'Completed': 10,
                'Review': 2,
                'Planning': 1
            },
            contentByType: {
                'blog_post': 20,
                'email': 12,
                'social_media': 8,
                'landing_page': 3,
                'copy': 2
            },
            monthlyActivity: [
                {
                    month: 'Jan 2024',
                    projectsStarted: 3,
                    projectsCompleted: 2,
                    contentDelivered: 8,
                    amountSpent: { amount: 2500.00, currency: 'USD' }
                },
                {
                    month: 'Feb 2024',
                    projectsStarted: 2,
                    projectsCompleted: 3,
                    contentDelivered: 12,
                    amountSpent: { amount: 3200.00, currency: 'USD' }
                },
                {
                    month: 'Mar 2024',
                    projectsStarted: 4,
                    projectsCompleted: 2,
                    contentDelivered: 10,
                    amountSpent: { amount: 2800.00, currency: 'USD' }
                },
                {
                    month: 'Apr 2024',
                    projectsStarted: 2,
                    projectsCompleted: 4,
                    contentDelivered: 15,
                    amountSpent: { amount: 4250.00, currency: 'USD' }
                }
            ],
            performanceMetrics: {
                onTimeDeliveryRate: 92.5,
                clientSatisfactionAvg: 4.6,
                revisionRequestRate: 15.2,
                averageResponseTime: '2.4 hours'
            },
            fromDate: this.formatDateForAPI(this.currentDateRange.from),
            toDate: this.formatDateForAPI(this.currentDateRange.to)
        };
    }

    getMockReport(reportType) {
        const reports = {
            'ProjectSummary': {
                reportId: 'report-1',
                title: 'Project Summary Report',
                summary: 'Comprehensive overview of all projects and their current status',
                insights: [
                    'Project completion rate has improved by 15% over the last quarter',
                    'SEO content projects show the highest client satisfaction ratings',
                    'Average project delivery time is 2.3 weeks'
                ],
                recommendations: [
                    'Consider expanding SEO content offerings based on high satisfaction',
                    'Implement project template to reduce setup time',
                    'Schedule regular check-ins for projects longer than 3 weeks'
                ],
                generatedAt: new Date().toISOString()
            },
            'ContentDelivery': {
                reportId: 'report-2',
                title: 'Content Delivery Performance',
                summary: 'Analysis of content delivery timelines and quality metrics',
                insights: [
                    'Blog posts have the fastest turnaround time at 3.2 days average',
                    'Email campaigns require 2.1 revisions on average',
                    'Landing pages achieve 95% first-time approval rate'
                ],
                recommendations: [
                    'Streamline email campaign briefing process',
                    'Create content templates for faster blog post creation',
                    'Leverage landing page success patterns for other content types'
                ],
                generatedAt: new Date().toISOString()
            }
        };

        return reports[reportType] || reports['ProjectSummary'];
    }
}