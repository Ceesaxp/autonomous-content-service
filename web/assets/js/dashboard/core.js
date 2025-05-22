// Dashboard Core Module - Handles common utilities and shared functionality

export class DashboardCore {
    constructor() {
        this.currentClientId = this.getClientIdFromUrl() || 'demo-client-id';
        this.apiBaseUrl = '/api/v1';
        this.eventBus = new EventTarget();
        this.cache = new Map();
        this.cacheTimeout = 5 * 60 * 1000; // 5 minutes
    }

    // Client ID Management
    getClientIdFromUrl() {
        const urlParams = new URLSearchParams(window.location.search);
        return urlParams.get('clientId') || localStorage.getItem('dashboardClientId');
    }

    setClientId(clientId) {
        this.currentClientId = clientId;
        localStorage.setItem('dashboardClientId', clientId);
        this.eventBus.dispatchEvent(new CustomEvent('clientChanged', { detail: { clientId } }));
    }

    // Event Bus
    on(event, callback) {
        this.eventBus.addEventListener(event, callback);
    }

    emit(event, data) {
        this.eventBus.dispatchEvent(new CustomEvent(event, { detail: data }));
    }

    // Cache Management
    setCache(key, data, timeout = this.cacheTimeout) {
        const expiry = Date.now() + timeout;
        this.cache.set(key, { data, expiry });
    }

    getCache(key) {
        const cached = this.cache.get(key);
        if (!cached) return null;
        
        if (Date.now() > cached.expiry) {
            this.cache.delete(key);
            return null;
        }
        
        return cached.data;
    }

    clearCache(pattern = null) {
        if (pattern) {
            for (const key of this.cache.keys()) {
                if (key.includes(pattern)) {
                    this.cache.delete(key);
                }
            }
        } else {
            this.cache.clear();
        }
    }

    // DOM Utilities
    $(selector, context = document) {
        return context.querySelector(selector);
    }

    $$(selector, context = document) {
        return Array.from(context.querySelectorAll(selector));
    }

    createElement(tag, attributes = {}, content = '') {
        const element = document.createElement(tag);
        
        Object.entries(attributes).forEach(([key, value]) => {
            if (key === 'className') {
                element.className = value;
            } else if (key.startsWith('data-')) {
                element.setAttribute(key, value);
            } else {
                element[key] = value;
            }
        });
        
        if (content) {
            if (typeof content === 'string') {
                element.innerHTML = content;
            } else {
                element.appendChild(content);
            }
        }
        
        return element;
    }

    // Loading States
    showLoading(element, message = 'Loading...') {
        const loader = this.createElement('div', { className: 'loading-overlay' }, `
            <div class="loading-spinner"></div>
            <p>${message}</p>
        `);
        
        element.style.position = 'relative';
        element.appendChild(loader);
        return loader;
    }

    hideLoading(element) {
        const loader = element.querySelector('.loading-overlay');
        if (loader) {
            loader.remove();
        }
    }

    // Toast Notifications
    showToast(message, type = 'info', duration = 5000) {
        const toast = this.createElement('div', {
            className: `toast toast-${type}`,
            'data-duration': duration
        }, `
            <div class="toast-content">
                <span class="toast-icon">${this.getToastIcon(type)}</span>
                <span class="toast-message">${message}</span>
                <button class="toast-close">&times;</button>
            </div>
        `);

        // Add toast styles if not already present
        this.ensureToastStyles();

        // Add to container
        let container = this.$('.toast-container');
        if (!container) {
            container = this.createElement('div', { className: 'toast-container' });
            document.body.appendChild(container);
        }

        container.appendChild(toast);

        // Auto remove
        setTimeout(() => {
            this.removeToast(toast);
        }, duration);

        // Close button handler
        toast.querySelector('.toast-close').addEventListener('click', () => {
            this.removeToast(toast);
        });

        return toast;
    }

    removeToast(toast) {
        toast.style.animation = 'slideOutRight 0.3s ease';
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 300);
    }

    getToastIcon(type) {
        const icons = {
            success: '✅',
            error: '❌',
            warning: '⚠️',
            info: 'ℹ️'
        };
        return icons[type] || icons.info;
    }

    ensureToastStyles() {
        if (this.$('#toast-styles')) return;

        const styles = this.createElement('style', { id: 'toast-styles' }, `
            .toast-container {
                position: fixed;
                top: 20px;
                right: 20px;
                z-index: 10000;
                display: flex;
                flex-direction: column;
                gap: 12px;
                max-width: 400px;
            }
            
            .toast {
                background: white;
                border: 1px solid #e2e8f0;
                border-radius: 8px;
                box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
                animation: slideInRight 0.3s ease;
                overflow: hidden;
            }
            
            .toast-success { border-left: 4px solid #10b981; }
            .toast-error { border-left: 4px solid #ef4444; }
            .toast-warning { border-left: 4px solid #f59e0b; }
            .toast-info { border-left: 4px solid #06b6d4; }
            
            .toast-content {
                display: flex;
                align-items: center;
                padding: 16px;
                gap: 12px;
            }
            
            .toast-icon {
                font-size: 18px;
                flex-shrink: 0;
            }
            
            .toast-message {
                flex: 1;
                font-size: 14px;
                line-height: 1.4;
            }
            
            .toast-close {
                background: none;
                border: none;
                font-size: 18px;
                cursor: pointer;
                color: #94a3b8;
                flex-shrink: 0;
                padding: 0;
                width: 20px;
                height: 20px;
                display: flex;
                align-items: center;
                justify-content: center;
            }
            
            .toast-close:hover {
                color: #475569;
            }
            
            @keyframes slideInRight {
                from { transform: translateX(100%); opacity: 0; }
                to { transform: translateX(0); opacity: 1; }
            }
            
            @keyframes slideOutRight {
                from { transform: translateX(0); opacity: 1; }
                to { transform: translateX(100%); opacity: 0; }
            }
            
            @media (max-width: 768px) {
                .toast-container {
                    left: 20px;
                    right: 20px;
                    max-width: none;
                }
            }
        `);

        document.head.appendChild(styles);
    }

    // Date Utilities
    formatDate(date, options = {}) {
        const defaultOptions = {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        };
        
        return new Intl.DateTimeFormat('en-US', { ...defaultOptions, ...options })
            .format(new Date(date));
    }

    formatTime(date) {
        return new Intl.DateTimeFormat('en-US', {
            hour: 'numeric',
            minute: '2-digit',
            hour12: true
        }).format(new Date(date));
    }

    formatDateTime(date) {
        return `${this.formatDate(date)} at ${this.formatTime(date)}`;
    }

    getRelativeTime(date) {
        const now = new Date();
        const target = new Date(date);
        const diffInSeconds = Math.floor((now - target) / 1000);

        if (diffInSeconds < 60) return 'Just now';
        if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
        if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
        if (diffInSeconds < 604800) return `${Math.floor(diffInSeconds / 86400)}d ago`;
        
        return this.formatDate(date);
    }

    getDaysFromNow(date) {
        const now = new Date();
        const target = new Date(date);
        const diffInMs = target - now;
        return Math.ceil(diffInMs / (1000 * 60 * 60 * 24));
    }

    // Currency Utilities
    formatCurrency(amount, currency = 'USD') {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency
        }).format(amount);
    }

    // String Utilities
    truncate(str, length = 100, suffix = '...') {
        if (str.length <= length) return str;
        return str.substring(0, length) + suffix;
    }

    slugify(str) {
        return str
            .toLowerCase()
            .trim()
            .replace(/[^\w\s-]/g, '')
            .replace(/[\s_-]+/g, '-')
            .replace(/^-+|-+$/g, '');
    }

    // Number Utilities
    formatNumber(num, decimals = 0) {
        return new Intl.NumberFormat('en-US', {
            minimumFractionDigits: decimals,
            maximumFractionDigits: decimals
        }).format(num);
    }

    formatPercentage(num, decimals = 1) {
        return new Intl.NumberFormat('en-US', {
            style: 'percent',
            minimumFractionDigits: decimals,
            maximumFractionDigits: decimals
        }).format(num / 100);
    }

    // Validation Utilities
    isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    isValidUrl(url) {
        try {
            new URL(url);
            return true;
        } catch {
            return false;
        }
    }

    // Local Storage Utilities
    setStorage(key, value, isSession = false) {
        try {
            const storage = isSession ? sessionStorage : localStorage;
            storage.setItem(key, JSON.stringify(value));
            return true;
        } catch (error) {
            console.error('Storage error:', error);
            return false;
        }
    }

    getStorage(key, isSession = false) {
        try {
            const storage = isSession ? sessionStorage : localStorage;
            const value = storage.getItem(key);
            return value ? JSON.parse(value) : null;
        } catch (error) {
            console.error('Storage error:', error);
            return null;
        }
    }

    removeStorage(key, isSession = false) {
        try {
            const storage = isSession ? sessionStorage : localStorage;
            storage.removeItem(key);
            return true;
        } catch (error) {
            console.error('Storage error:', error);
            return false;
        }
    }

    // Debounce Utility
    debounce(func, wait, immediate = false) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                timeout = null;
                if (!immediate) func(...args);
            };
            const callNow = immediate && !timeout;
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
            if (callNow) func(...args);
        };
    }

    // Throttle Utility
    throttle(func, limit) {
        let inThrottle;
        return function(...args) {
            if (!inThrottle) {
                func.apply(this, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }

    // Error Handling
    handleError(error, context = 'Dashboard') {
        console.error(`[${context}] Error:`, error);
        
        let message = 'An unexpected error occurred. Please try again.';
        
        if (error.message) {
            message = error.message;
        } else if (typeof error === 'string') {
            message = error;
        }
        
        this.showToast(message, 'error');
        this.emit('error', { error, context, message });
    }

    // Success Handling
    handleSuccess(message, data = null) {
        this.showToast(message, 'success');
        this.emit('success', { message, data });
    }

    // Initialize core functionality
    init() {
        this.initKeyboardShortcuts();
        this.initErrorHandling();
        this.initOfflineDetection();
        console.log('Dashboard Core initialized');
    }

    initKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            // Ctrl/Cmd + K for search
            if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
                e.preventDefault();
                this.emit('search-focus');
            }
            
            // Escape to close modals
            if (e.key === 'Escape') {
                this.emit('escape-pressed');
            }
        });
    }

    initErrorHandling() {
        window.addEventListener('unhandledrejection', (event) => {
            this.handleError(event.reason, 'Unhandled Promise');
        });

        window.addEventListener('error', (event) => {
            this.handleError(event.error, 'Global Error');
        });
    }

    initOfflineDetection() {
        window.addEventListener('online', () => {
            this.showToast('Connection restored', 'success');
            this.emit('online');
        });

        window.addEventListener('offline', () => {
            this.showToast('Connection lost. Working offline.', 'warning');
            this.emit('offline');
        });
    }
}

// Export singleton instance
export const core = new DashboardCore();