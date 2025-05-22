// Main JavaScript functionality for Autonomous Content Service

class WebsiteController {
    constructor() {
        this.currentStep = 1;
        this.maxSteps = 4;
        this.priceCalculator = new PriceCalculator();
        this.portfolio = new PortfolioManager();
        this.analytics = new AnalyticsManager();
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.initializeComponents();
        this.trackPageView();
    }

    setupEventListeners() {
        // Navigation toggle for mobile
        const navToggle = document.querySelector('.nav-toggle');
        const navMenu = document.querySelector('.nav-menu');
        
        if (navToggle && navMenu) {
            navToggle.addEventListener('click', () => {
                navMenu.classList.toggle('active');
            });
        }

        // Form handling
        const projectForm = document.getElementById('project-form');
        if (projectForm) {
            projectForm.addEventListener('submit', this.handleFormSubmit.bind(this));
            
            // Auto-calculate pricing as user types
            const inputs = projectForm.querySelectorAll('input, select, textarea');
            inputs.forEach(input => {
                input.addEventListener('change', this.updatePriceEstimate.bind(this));
            });
        }

        // Portfolio filters
        const filterTabs = document.querySelectorAll('.filter-tab');
        filterTabs.forEach(tab => {
            tab.addEventListener('click', (e) => {
                this.portfolio.filterItems(e.target.dataset.filter);
                this.updateActiveFilter(e.target);
            });
        });

        // Price calculator modal
        const priceForm = document.getElementById('price-form');
        if (priceForm) {
            const inputs = priceForm.querySelectorAll('input, select');
            inputs.forEach(input => {
                input.addEventListener('change', this.calculatePrice.bind(this));
            });
        }

        // Smooth scrolling for anchor links
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function (e) {
                e.preventDefault();
                const target = document.querySelector(this.getAttribute('href'));
                if (target) {
                    target.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });

        // Track button clicks
        document.querySelectorAll('.btn-primary, .btn-secondary, .btn-outline').forEach(button => {
            button.addEventListener('click', (e) => {
                this.analytics.trackEvent('button_click', {
                    button_text: e.target.textContent.trim(),
                    button_type: e.target.className,
                    page: window.location.pathname
                });
            });
        });
    }

    initializeComponents() {
        // Initialize pricing display
        this.updatePriceEstimate();
        
        // Initialize portfolio
        this.portfolio.init();
        
        // Initialize analytics
        this.analytics.init();
        
        // Auto-populate form fields from URL parameters
        this.populateFormFromURL();
    }

    populateFormFromURL() {
        const urlParams = new URLSearchParams(window.location.search);
        const service = urlParams.get('service');
        
        if (service) {
            const serviceSelect = document.getElementById('project-type');
            if (serviceSelect) {
                serviceSelect.value = service;
                this.updatePriceEstimate();
            }
        }
    }

    // Form Step Navigation
    nextStep() {
        if (this.validateCurrentStep()) {
            this.currentStep = Math.min(this.currentStep + 1, this.maxSteps);
            this.updateFormDisplay();
            this.updateProgressBar();
            this.analytics.trackEvent('form_step_completed', {
                step: this.currentStep - 1,
                form_type: 'project_request'
            });
        }
    }

    prevStep() {
        this.currentStep = Math.max(this.currentStep - 1, 1);
        this.updateFormDisplay();
        this.updateProgressBar();
    }

    updateFormDisplay() {
        document.querySelectorAll('.form-step').forEach((step, index) => {
            step.classList.toggle('active', index + 1 === this.currentStep);
        });
    }

    updateProgressBar() {
        const progressFill = document.querySelector('.progress-fill');
        const progressSteps = document.querySelectorAll('.progress-steps .step');
        
        if (progressFill) {
            const percentage = (this.currentStep / this.maxSteps) * 100;
            progressFill.style.width = `${percentage}%`;
        }
        
        progressSteps.forEach((step, index) => {
            step.classList.toggle('active', index < this.currentStep);
        });
    }

    validateCurrentStep() {
        const currentStepElement = document.querySelector(`.form-step[data-step="${this.currentStep}"]`);
        if (!currentStepElement) return false;
        
        const requiredInputs = currentStepElement.querySelectorAll('input[required], select[required], textarea[required]');
        
        for (let input of requiredInputs) {
            if (!input.value.trim()) {
                input.focus();
                this.showValidationError(input, 'This field is required');
                return false;
            }
        }
        
        return true;
    }

    showValidationError(input, message) {
        // Remove existing error messages
        const existingError = input.parentNode.querySelector('.error-message');
        if (existingError) {
            existingError.remove();
        }
        
        // Add new error message
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.textContent = message;
        errorDiv.style.color = 'var(--error-color)';
        errorDiv.style.fontSize = '0.875rem';
        errorDiv.style.marginTop = '0.25rem';
        
        input.parentNode.appendChild(errorDiv);
        input.style.borderColor = 'var(--error-color)';
        
        // Remove error styling on input
        input.addEventListener('input', () => {
            input.style.borderColor = '';
            if (errorDiv.parentNode) {
                errorDiv.remove();
            }
        }, { once: true });
    }

    async handleFormSubmit(e) {
        e.preventDefault();
        
        if (!this.validateCurrentStep()) {
            return;
        }
        
        try {
            const formData = this.collectFormData();
            
            // Show loading state
            const submitButton = e.target.querySelector('button[type="submit"]');
            const originalText = submitButton.textContent;
            submitButton.textContent = 'Submitting...';
            submitButton.disabled = true;
            
            // Simulate API call to backend
            await this.submitProjectRequest(formData);
            
            // Move to success step
            this.currentStep = this.maxSteps;
            this.updateFormDisplay();
            this.updateProgressBar();
            
            this.analytics.trackEvent('form_submitted', {
                form_type: 'project_request',
                project_type: formData.projectType,
                estimated_value: formData.estimatedPrice
            });
            
        } catch (error) {
            console.error('Form submission failed:', error);
            this.showFormError('Failed to submit request. Please try again.');
            
            // Reset button
            const submitButton = e.target.querySelector('button[type="submit"]');
            submitButton.textContent = originalText;
            submitButton.disabled = false;
        }
    }

    collectFormData() {
        const form = document.getElementById('project-form');
        const formData = new FormData(form);
        const data = {};
        
        // Collect all form values
        ['project-type', 'project-description', 'word-count', 'deadline', 
         'tone', 'target-audience', 'keywords', 'references',
         'client-name', 'client-email', 'company', 'phone', 'budget'].forEach(id => {
            const element = document.getElementById(id);
            if (element) {
                data[this.camelCase(id)] = element.value;
            }
        });
        
        // Add calculated price
        data.estimatedPrice = this.priceCalculator.getCurrentPrice();
        data.timestamp = new Date().toISOString();
        
        return data;
    }

    async submitProjectRequest(data) {
        // In a real implementation, this would call the Go backend API
        // For now, simulate an API call
        return new Promise((resolve) => {
            setTimeout(() => {
                console.log('Project request submitted:', data);
                resolve(data);
            }, 2000);
        });
    }

    showFormError(message) {
        const errorDiv = document.createElement('div');
        errorDiv.className = 'form-error';
        errorDiv.textContent = message;
        errorDiv.style.background = 'var(--error-color)';
        errorDiv.style.color = 'white';
        errorDiv.style.padding = 'var(--spacing-md)';
        errorDiv.style.borderRadius = 'var(--radius-md)';
        errorDiv.style.marginBottom = 'var(--spacing-md)';
        
        const form = document.getElementById('project-form');
        form.insertBefore(errorDiv, form.firstChild);
        
        setTimeout(() => {
            errorDiv.remove();
        }, 5000);
    }

    updatePriceEstimate() {
        const estimate = this.priceCalculator.estimatePrice();
        const priceDisplay = document.getElementById('price-estimate');
        
        if (priceDisplay) {
            const priceAmount = priceDisplay.querySelector('.price-amount');
            if (priceAmount && estimate > 0) {
                priceAmount.textContent = `$${estimate}`;
            }
        }
    }

    calculatePrice() {
        const price = this.priceCalculator.calculateDetailedPrice();
        const priceElement = document.getElementById('calculated-price');
        
        if (priceElement) {
            priceElement.textContent = `$${price}`;
        }
    }

    updateActiveFilter(activeTab) {
        document.querySelectorAll('.filter-tab').forEach(tab => {
            tab.classList.remove('active');
        });
        activeTab.classList.add('active');
    }

    trackPageView() {
        this.analytics.trackPageView();
    }

    camelCase(str) {
        return str.replace(/-([a-z])/g, (g) => g[1].toUpperCase());
    }
}

class PriceCalculator {
    constructor() {
        this.basePrices = {
            'Blog Posts & Articles': 0.08,
            'Marketing Copy': 0.12,
            'Technical Documentation': 0.15,
            'Social Media Content': 0.05,
            'Email Marketing': 0.10
        };
        
        this.multipliers = {
            urgency: {
                'rush': 1.5,
                'express': 1.25,
                'standard': 1.0,
                'flexible': 0.9
            },
            complexity: {
                'basic': 1.0,
                'intermediate': 1.2,
                'advanced': 1.4,
                'expert': 1.6
            }
        };
    }

    estimatePrice() {
        const projectType = this.getFieldValue('project-type');
        const wordCount = this.parseWordCount(this.getFieldValue('word-count'));
        
        if (!projectType || !wordCount) return 0;
        
        const basePrice = this.basePrices[projectType] || 0.08;
        return Math.round(basePrice * wordCount);
    }

    calculateDetailedPrice() {
        const projectType = this.getFieldValue('content-type');
        const wordCount = parseInt(this.getFieldValue('word-count')) || 500;
        const urgency = this.getFieldValue('urgency') || 'standard';
        const complexity = this.getFieldValue('complexity') || 'basic';
        
        if (!projectType) return 0;
        
        const basePrice = this.basePrices[projectType] || 0.08;
        const urgencyMultiplier = this.multipliers.urgency[urgency] || 1.0;
        const complexityMultiplier = this.multipliers.complexity[complexity] || 1.0;
        
        const finalPrice = basePrice * wordCount * urgencyMultiplier * complexityMultiplier;
        return Math.round(finalPrice);
    }

    getCurrentPrice() {
        return this.estimatePrice();
    }

    parseWordCount(wordCountStr) {
        if (!wordCountStr) return 0;
        
        const ranges = {
            '100-500': 300,
            '500-1000': 750,
            '1000-2000': 1500,
            '2000-5000': 3500,
            '5000+': 7500
        };
        
        return ranges[wordCountStr] || 500;
    }

    getFieldValue(fieldId) {
        const element = document.getElementById(fieldId);
        return element ? element.value : '';
    }
}

class PortfolioManager {
    constructor() {
        this.items = [];
        this.currentFilter = 'all';
    }

    init() {
        this.loadPortfolioItems();
    }

    loadPortfolioItems() {
        // In a real implementation, this would fetch from the API
        // For now, items are statically defined in HTML
        this.items = Array.from(document.querySelectorAll('.portfolio-item'));
    }

    filterItems(category) {
        this.currentFilter = category;
        
        this.items.forEach(item => {
            const itemCategory = item.dataset.category;
            const shouldShow = category === 'all' || itemCategory === category;
            
            item.style.display = shouldShow ? 'block' : 'none';
        });
    }
}

class AnalyticsManager {
    constructor() {
        this.sessionId = this.generateSessionId();
        this.events = [];
    }

    init() {
        this.trackSession();
    }

    trackPageView() {
        this.trackEvent('page_view', {
            page: window.location.pathname,
            referrer: document.referrer,
            timestamp: new Date().toISOString()
        });
    }

    trackEvent(eventName, data = {}) {
        const event = {
            event: eventName,
            session_id: this.sessionId,
            timestamp: new Date().toISOString(),
            url: window.location.href,
            user_agent: navigator.userAgent,
            ...data
        };
        
        this.events.push(event);
        
        // In a real implementation, send to analytics service
        console.log('Analytics event:', event);
        
        // Store locally for now
        this.storeEvent(event);
    }

    trackSession() {
        this.trackEvent('session_start', {
            screen_resolution: `${screen.width}x${screen.height}`,
            viewport_size: `${window.innerWidth}x${window.innerHeight}`,
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
        });
    }

    storeEvent(event) {
        try {
            const events = JSON.parse(localStorage.getItem('analytics_events') || '[]');
            events.push(event);
            
            // Keep only last 100 events
            if (events.length > 100) {
                events.splice(0, events.length - 100);
            }
            
            localStorage.setItem('analytics_events', JSON.stringify(events));
        } catch (error) {
            console.error('Failed to store analytics event:', error);
        }
    }

    generateSessionId() {
        return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }
}

// Global functions for HTML onclick handlers
window.nextStep = function() {
    if (window.websiteController) {
        window.websiteController.nextStep();
    }
};

window.prevStep = function() {
    if (window.websiteController) {
        window.websiteController.prevStep();
    }
};

window.openPriceCalculator = function(service) {
    const modal = document.getElementById('pricing-calculator');
    const serviceSelect = document.getElementById('content-type');
    
    if (modal) {
        modal.style.display = 'block';
        if (serviceSelect && service) {
            serviceSelect.value = service;
        }
        
        if (window.websiteController) {
            window.websiteController.calculatePrice();
        }
    }
};

window.closePriceCalculator = function() {
    const modal = document.getElementById('pricing-calculator');
    if (modal) {
        modal.style.display = 'none';
    }
};

window.proceedWithQuote = function() {
    const service = document.getElementById('content-type').value;
    window.location.href = `contact.html${service ? '?service=' + encodeURIComponent(service) : ''}`;
};

window.viewPortfolioItem = function(itemId) {
    // Implementation for viewing portfolio items
    console.log('Viewing portfolio item:', itemId);
};

window.closePortfolioModal = function() {
    const modal = document.getElementById('portfolio-modal');
    if (modal) {
        modal.style.display = 'none';
    }
};

window.requestSimilarContent = function() {
    window.location.href = 'contact.html';
};

window.openClientPortal = function() {
    // In a real implementation, this would open the client portal
    alert('Client portal functionality will be implemented in Phase 4.2');
};

// Initialize the website when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.websiteController = new WebsiteController();
});