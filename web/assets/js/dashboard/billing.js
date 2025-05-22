// Dashboard Billing Module

import { core } from './core.js';
import { api } from './api.js';

export class BillingManager {
    constructor() {
        this.billingHistory = [];
        this.outstandingInvoices = [];
        this.billingStats = {
            totalSpent: { amount: 0, currency: 'USD' },
            outstandingAmount: { amount: 0, currency: 'USD' },
            nextPayment: null
        };
    }

    async init() {
        this.setupEventListeners();
        await this.loadData();
    }

    setupEventListeners() {
        // Any billing-specific event listeners would go here
        // For now, just setup refresh handlers
        core.on('billing-refresh', () => {
            this.loadData();
        });
    }

    async loadData() {
        try {
            await Promise.all([
                this.loadBillingHistory(),
                this.loadOutstandingInvoices()
            ]);
            
            this.calculateBillingStats();
            this.renderBillingData();
            
        } catch (error) {
            console.error('Error loading billing data:', error);
            // Use mock data for development
            this.loadMockBillingData();
            this.renderBillingData();
        }
    }

    async loadBillingHistory() {
        try {
            this.billingHistory = await api.getBillingHistory(core.currentClientId, 50, 0);
        } catch (error) {
            this.billingHistory = this.getMockBillingHistory();
        }
    }

    async loadOutstandingInvoices() {
        try {
            this.outstandingInvoices = await api.getOutstandingInvoices(core.currentClientId);
        } catch (error) {
            this.outstandingInvoices = this.getMockOutstandingInvoices();
        }
    }

    loadMockBillingData() {
        this.billingHistory = this.getMockBillingHistory();
        this.outstandingInvoices = this.getMockOutstandingInvoices();
        this.calculateBillingStats();
    }

    calculateBillingStats() {
        // Calculate total spent
        const totalSpent = this.billingHistory
            .filter(item => item.status === 'Completed')
            .reduce((sum, item) => sum + item.amount.amount, 0);

        // Calculate outstanding amount
        const outstandingAmount = this.outstandingInvoices
            .reduce((sum, invoice) => sum + invoice.amount.amount, 0);

        // Find next payment due date
        const nextPayment = this.outstandingInvoices
            .filter(invoice => invoice.status === 'Pending')
            .sort((a, b) => new Date(a.dueDate) - new Date(b.dueDate))[0];

        this.billingStats = {
            totalSpent: { amount: totalSpent, currency: 'USD' },
            outstandingAmount: { amount: outstandingAmount, currency: 'USD' },
            nextPayment: nextPayment ? nextPayment.dueDate : null
        };
    }

    renderBillingData() {
        this.renderBillingStats();
        this.renderOutstandingInvoices();
        this.renderBillingHistory();
    }

    renderBillingStats() {
        // Update billing summary stats
        const totalSpentElement = core.$('#totalSpent');
        const outstandingAmountElement = core.$('#outstandingAmount');
        const nextPaymentElement = core.$('#nextPayment');

        if (totalSpentElement) {
            totalSpentElement.textContent = core.formatCurrency(this.billingStats.totalSpent.amount);
        }

        if (outstandingAmountElement) {
            outstandingAmountElement.textContent = core.formatCurrency(this.billingStats.outstandingAmount.amount);
        }

        if (nextPaymentElement) {
            nextPaymentElement.textContent = this.billingStats.nextPayment ? 
                core.formatDate(this.billingStats.nextPayment) : 
                'No pending payments';
        }
    }

    renderOutstandingInvoices() {
        const container = core.$('#outstandingInvoicesList');
        if (!container) return;

        container.innerHTML = '';

        if (this.outstandingInvoices.length === 0) {
            container.innerHTML = `
                <div class="no-outstanding">
                    <p>No outstanding invoices</p>
                    <p class="text-success">All caught up! 🎉</p>
                </div>
            `;
            return;
        }

        this.outstandingInvoices.forEach(invoice => {
            const invoiceElement = this.createInvoiceElement(invoice, true);
            container.appendChild(invoiceElement);
        });
    }

    renderBillingHistory() {
        const container = core.$('#billingHistoryList');
        if (!container) return;

        container.innerHTML = '';

        if (this.billingHistory.length === 0) {
            container.innerHTML = `
                <div class="no-history">
                    <p>No billing history available</p>
                </div>
            `;
            return;
        }

        // Show recent 10 items in the dashboard view
        const recentHistory = this.billingHistory.slice(0, 10);
        
        recentHistory.forEach(item => {
            const historyElement = this.createBillingHistoryElement(item);
            container.appendChild(historyElement);
        });

        // Add "View All" link if there are more items
        if (this.billingHistory.length > 10) {
            const viewAllElement = core.createElement('div', { className: 'view-all-billing' }, `
                <button class="btn btn-link" onclick="this.showAllBillingHistory()">
                    View All ${this.billingHistory.length} Items
                </button>
            `);
            container.appendChild(viewAllElement);
        }
    }

    createInvoiceElement(invoice, isOutstanding = false) {
        const element = core.createElement('div', {
            className: `invoice-item ${isOutstanding ? 'outstanding' : ''}`,
            'data-invoice-id': invoice.billingId
        });

        const dueDate = new Date(invoice.dueDate);
        const isOverdue = dueDate < new Date() && invoice.status === 'Pending';
        const urgencyClass = isOverdue ? 'overdue' : 
                            this.getDaysUntilDue(dueDate) <= 3 ? 'urgent' : '';

        element.innerHTML = `
            <div class="invoice-details">
                <div class="invoice-header">
                    <span class="invoice-number">${invoice.invoiceNumber}</span>
                    <span class="invoice-status status-${invoice.status.toLowerCase()}">${invoice.status}</span>
                </div>
                <div class="invoice-description">${invoice.description}</div>
                <div class="invoice-meta">
                    <span class="invoice-date">Due: ${core.formatDate(invoice.dueDate)}</span>
                    ${isOverdue ? '<span class="overdue-notice">Overdue</span>' : ''}
                </div>
            </div>
            <div class="invoice-amount ${urgencyClass}">
                ${core.formatCurrency(invoice.amount.amount)}
            </div>
        `;

        // Add click handler for outstanding invoices
        if (isOutstanding && invoice.status === 'Pending') {
            element.style.cursor = 'pointer';
            element.addEventListener('click', () => {
                this.showInvoiceDetails(invoice);
            });
        }

        return element;
    }

    createBillingHistoryElement(item) {
        const element = core.createElement('div', {
            className: 'billing-item',
            'data-billing-id': item.billingId
        });

        const statusIcon = this.getStatusIcon(item.status);
        const date = item.paidDate || item.createdAt;

        element.innerHTML = `
            <div class="billing-details">
                <div class="billing-header">
                    <span class="billing-description">${item.description}</span>
                    <span class="billing-status">${statusIcon} ${item.status}</span>
                </div>
                <div class="billing-meta">
                    <span class="billing-invoice">${item.invoiceNumber}</span>
                    <span class="billing-date">${core.formatDate(date)}</span>
                </div>
            </div>
            <div class="billing-amount">
                ${core.formatCurrency(item.amount.amount)}
            </div>
        `;

        return element;
    }

    showInvoiceDetails(invoice) {
        // Create a simple invoice details modal
        const modalContent = `
            <div class="invoice-details-modal">
                <h3>Invoice Details</h3>
                
                <div class="invoice-info">
                    <div class="info-row">
                        <label>Invoice Number:</label>
                        <span>${invoice.invoiceNumber}</span>
                    </div>
                    <div class="info-row">
                        <label>Amount:</label>
                        <span class="amount">${core.formatCurrency(invoice.amount.amount)}</span>
                    </div>
                    <div class="info-row">
                        <label>Due Date:</label>
                        <span>${core.formatDate(invoice.dueDate)}</span>
                    </div>
                    <div class="info-row">
                        <label>Status:</label>
                        <span class="status-${invoice.status.toLowerCase()}">${invoice.status}</span>
                    </div>
                    <div class="info-row">
                        <label>Description:</label>
                        <span>${invoice.description}</span>
                    </div>
                </div>

                ${invoice.lineItems && invoice.lineItems.length > 0 ? `
                    <div class="line-items">
                        <h4>Line Items</h4>
                        <div class="items-list">
                            ${invoice.lineItems.map(item => `
                                <div class="line-item">
                                    <span class="item-description">${item.description}</span>
                                    <span class="item-quantity">${item.quantity}x</span>
                                    <span class="item-price">${core.formatCurrency(item.total.amount)}</span>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                ` : ''}

                <div class="invoice-actions">
                    <button class="btn btn-secondary" onclick="this.closeInvoiceDetails()">Close</button>
                    ${invoice.status === 'Pending' ? `
                        <button class="btn btn-primary" onclick="this.processPayment('${invoice.billingId}')">
                            Pay Now
                        </button>
                    ` : ''}
                </div>
            </div>
        `;

        // Show modal (simplified implementation)
        core.showToast('Invoice details would be displayed in a modal', 'info');
    }

    getDaysUntilDue(dueDate) {
        const now = new Date();
        const due = new Date(dueDate);
        const diffTime = due - now;
        return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    }

    getStatusIcon(status) {
        const icons = {
            'Pending': '⏳',
            'Completed': '✅',
            'Overdue': '❌',
            'Cancelled': '🚫'
        };
        return icons[status] || '📄';
    }

    showAllBillingHistory() {
        // This would typically open a full billing history view
        // For now, just show a message
        core.showToast('Full billing history view would be implemented here', 'info');
    }

    async processPayment(billingId) {
        try {
            // This would integrate with a payment processor
            core.showToast('Payment processing would be implemented here', 'info');
            
            // For demo purposes, simulate successful payment
            setTimeout(() => {
                core.handleSuccess('Payment processed successfully');
                this.loadData(); // Refresh data
            }, 2000);
            
        } catch (error) {
            core.handleError(error, 'Process Payment');
        }
    }

    // Mock data methods
    getMockBillingHistory() {
        return [
            {
                billingId: 'bill-1',
                projectId: 'proj-123',
                clientId: core.currentClientId,
                invoiceNumber: 'INV-2024-001',
                amount: { amount: 2500.00, currency: 'USD' },
                status: 'Completed',
                dueDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
                paidDate: new Date(Date.now() - 25 * 24 * 60 * 60 * 1000).toISOString(),
                description: 'SEO Content Package - Blog Posts',
                lineItems: [
                    {
                        description: 'Blog Post - SEO Best Practices',
                        quantity: 1,
                        unitPrice: { amount: 500.00, currency: 'USD' },
                        total: { amount: 500.00, currency: 'USD' }
                    },
                    {
                        description: 'Blog Post - Technical SEO',
                        quantity: 1,
                        unitPrice: { amount: 500.00, currency: 'USD' },
                        total: { amount: 500.00, currency: 'USD' }
                    },
                    {
                        description: 'Landing Page Copy',
                        quantity: 3,
                        unitPrice: { amount: 500.00, currency: 'USD' },
                        total: { amount: 1500.00, currency: 'USD' }
                    }
                ],
                createdAt: new Date(Date.now() - 35 * 24 * 60 * 60 * 1000).toISOString()
            },
            {
                billingId: 'bill-2',
                projectId: 'proj-124',
                clientId: core.currentClientId,
                invoiceNumber: 'INV-2024-002',
                amount: { amount: 1800.00, currency: 'USD' },
                status: 'Completed',
                dueDate: new Date(Date.now() - 15 * 24 * 60 * 60 * 1000).toISOString(),
                paidDate: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000).toISOString(),
                description: 'Social Media Strategy Package',
                createdAt: new Date(Date.now() - 20 * 24 * 60 * 60 * 1000).toISOString()
            },
            {
                billingId: 'bill-3',
                projectId: 'proj-125',
                clientId: core.currentClientId,
                invoiceNumber: 'INV-2024-003',
                amount: { amount: 1200.00, currency: 'USD' },
                status: 'Completed',
                dueDate: new Date(Date.now() - 45 * 24 * 60 * 60 * 1000).toISOString(),
                paidDate: new Date(Date.now() - 40 * 24 * 60 * 60 * 1000).toISOString(),
                description: 'Email Marketing Campaign Setup',
                createdAt: new Date(Date.now() - 50 * 24 * 60 * 60 * 1000).toISOString()
            }
        ];
    }

    getMockOutstandingInvoices() {
        return [
            {
                billingId: 'bill-4',
                projectId: 'proj-123',
                clientId: core.currentClientId,
                invoiceNumber: 'INV-2024-004',
                amount: { amount: 1250.00, currency: 'USD' },
                status: 'Pending',
                dueDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
                description: 'SEO Content Package - Phase 2',
                lineItems: [
                    {
                        description: 'Blog Post - Link Building Guide',
                        quantity: 1,
                        unitPrice: { amount: 500.00, currency: 'USD' },
                        total: { amount: 500.00, currency: 'USD' }
                    },
                    {
                        description: 'Product Page Copy',
                        quantity: 3,
                        unitPrice: { amount: 250.00, currency: 'USD' },
                        total: { amount: 750.00, currency: 'USD' }
                    }
                ],
                createdAt: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString()
            },
            {
                billingId: 'bill-5',
                projectId: 'proj-126',
                clientId: core.currentClientId,
                invoiceNumber: 'INV-2024-005',
                amount: { amount: 950.00, currency: 'USD' },
                status: 'Pending',
                dueDate: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString(),
                description: 'Content Audit and Strategy',
                createdAt: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString()
            }
        ];
    }
}