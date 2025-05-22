// Analytics and tracking for autonomous web presence optimization

class AutonomousAnalytics {
    constructor() {
        this.sessionId = this.generateSessionId();
        this.visitorId = this.getOrCreateVisitorId();
        this.pageMetrics = new Map();
        this.heatmapData = [];
        this.performanceMetrics = {};
        
        this.init();
    }

    init() {
        this.startPerformanceTracking();
        this.setupEventListeners();
        this.startHeatmapTracking();
        this.trackPageLoad();
    }

    // Performance Monitoring
    startPerformanceTracking() {
        if ('performance' in window) {
            window.addEventListener('load', () => {
                setTimeout(() => {
                    this.collectPerformanceMetrics();
                }, 1000);
            });
        }
    }

    collectPerformanceMetrics() {
        const navigation = performance.getEntriesByType('navigation')[0];
        const paintEntries = performance.getEntriesByType('paint');
        
        this.performanceMetrics = {
            pageLoadTime: navigation.loadEventEnd - navigation.fetchStart,
            domContentLoaded: navigation.domContentLoadedEventEnd - navigation.fetchStart,
            firstContentfulPaint: paintEntries.find(entry => entry.name === 'first-contentful-paint')?.startTime || 0,
            timeToInteractive: navigation.domInteractive - navigation.fetchStart,
            resourceLoadTime: navigation.loadEventEnd - navigation.responseEnd,
            dnslookupTime: navigation.domainLookupEnd - navigation.domainLookupStart,
            serverResponseTime: navigation.responseEnd - navigation.requestStart
        };

        this.sendAnalyticsEvent('performance_metrics', this.performanceMetrics);
    }

    // User Behavior Tracking
    setupEventListeners() {
        // Track clicks
        document.addEventListener('click', (e) => {
            this.trackClick(e);
        });

        // Track scroll behavior
        let scrollTimeout;
        window.addEventListener('scroll', () => {
            clearTimeout(scrollTimeout);
            scrollTimeout = setTimeout(() => {
                this.trackScroll();
            }, 100);
        });

        // Track form interactions
        document.addEventListener('focus', (e) => {
            if (e.target.matches('input, textarea, select')) {
                this.trackFormInteraction('focus', e.target);
            }
        }, true);

        document.addEventListener('blur', (e) => {
            if (e.target.matches('input, textarea, select')) {
                this.trackFormInteraction('blur', e.target);
            }
        }, true);

        // Track page visibility changes
        document.addEventListener('visibilitychange', () => {
            this.trackVisibilityChange();
        });

        // Track page unload
        window.addEventListener('beforeunload', () => {
            this.trackPageUnload();
        });
    }

    trackClick(event) {
        const element = event.target;
        const clickData = {
            event_type: 'click',
            element_tag: element.tagName.toLowerCase(),
            element_id: element.id || null,
            element_class: element.className || null,
            element_text: element.textContent.trim().substring(0, 100),
            x_position: event.clientX,
            y_position: event.clientY,
            page_x: event.pageX,
            page_y: event.pageY,
            timestamp: Date.now()
        };

        // Track specific button types
        if (element.matches('.btn-primary, .btn-secondary, .btn-outline')) {
            clickData.button_type = this.getButtonType(element);
            clickData.cta_text = element.textContent.trim();
        }

        // Track navigation clicks
        if (element.matches('a')) {
            clickData.link_href = element.href;
            clickData.is_external = !element.href.includes(window.location.hostname);
        }

        this.sendAnalyticsEvent('user_click', clickData);
        this.addToHeatmap(event.pageX, event.pageY, 'click');
    }

    trackScroll() {
        const scrollPercent = Math.round((window.scrollY / (document.documentElement.scrollHeight - window.innerHeight)) * 100);
        const viewportHeight = window.innerHeight;
        const documentHeight = document.documentElement.scrollHeight;
        
        const scrollData = {
            event_type: 'scroll',
            scroll_percent: Math.min(scrollPercent, 100),
            scroll_position: window.scrollY,
            viewport_height: viewportHeight,
            document_height: documentHeight,
            timestamp: Date.now()
        };

        // Track milestone scrolls
        if (scrollPercent % 25 === 0 && scrollPercent > 0) {
            scrollData.milestone = `${scrollPercent}%`;
        }

        this.sendAnalyticsEvent('scroll_tracking', scrollData);
    }

    trackFormInteraction(action, element) {
        const formData = {
            event_type: 'form_interaction',
            action: action,
            field_type: element.type || element.tagName.toLowerCase(),
            field_id: element.id || null,
            field_name: element.name || null,
            form_id: element.form?.id || null,
            timestamp: Date.now()
        };

        this.sendAnalyticsEvent('form_interaction', formData);
    }

    trackVisibilityChange() {
        const visibilityData = {
            event_type: 'visibility_change',
            visibility_state: document.visibilityState,
            hidden: document.hidden,
            timestamp: Date.now()
        };

        this.sendAnalyticsEvent('visibility_change', visibilityData);
    }

    trackPageLoad() {
        const pageData = {
            event_type: 'page_load',
            url: window.location.href,
            path: window.location.pathname,
            referrer: document.referrer,
            user_agent: navigator.userAgent,
            screen_resolution: `${screen.width}x${screen.height}`,
            viewport_size: `${window.innerWidth}x${window.innerHeight}`,
            color_depth: screen.colorDepth,
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
            language: navigator.language,
            timestamp: Date.now()
        };

        this.sendAnalyticsEvent('page_load', pageData);
        this.startSessionTimer();
    }

    trackPageUnload() {
        const unloadData = {
            event_type: 'page_unload',
            session_duration: Date.now() - this.sessionStartTime,
            timestamp: Date.now()
        };

        // Send immediately
        this.sendAnalyticsEvent('page_unload', unloadData, true);
        this.saveHeatmapData();
    }

    // Heatmap Tracking
    startHeatmapTracking() {
        this.heatmapData = [];
        
        // Track mouse movements (throttled)
        let mouseMoveTimeout;
        document.addEventListener('mousemove', (e) => {
            clearTimeout(mouseMoveTimeout);
            mouseMoveTimeout = setTimeout(() => {
                this.addToHeatmap(e.pageX, e.pageY, 'move');
            }, 100);
        });
    }

    addToHeatmap(x, y, type) {
        this.heatmapData.push({
            x: x,
            y: y,
            type: type,
            timestamp: Date.now()
        });

        // Limit heatmap data size
        if (this.heatmapData.length > 1000) {
            this.heatmapData = this.heatmapData.slice(-500);
        }
    }

    saveHeatmapData() {
        if (this.heatmapData.length > 0) {
            const heatmapSummary = {
                event_type: 'heatmap_data',
                page_url: window.location.href,
                data_points: this.heatmapData.length,
                viewport_size: `${window.innerWidth}x${window.innerHeight}`,
                session_id: this.sessionId,
                timestamp: Date.now()
            };

            this.sendAnalyticsEvent('heatmap_summary', heatmapSummary, true);
        }
    }

    // A/B Testing Support
    initializeABTest(testName, variants) {
        const existingVariant = localStorage.getItem(`ab_test_${testName}`);
        
        if (existingVariant && variants.includes(existingVariant)) {
            return existingVariant;
        }

        const selectedVariant = variants[Math.floor(Math.random() * variants.length)];
        localStorage.setItem(`ab_test_${testName}`, selectedVariant);

        this.sendAnalyticsEvent('ab_test_assignment', {
            test_name: testName,
            variant: selectedVariant,
            timestamp: Date.now()
        });

        return selectedVariant;
    }

    trackABTestConversion(testName, conversionType, value = null) {
        const variant = localStorage.getItem(`ab_test_${testName}`);
        
        if (variant) {
            this.sendAnalyticsEvent('ab_test_conversion', {
                test_name: testName,
                variant: variant,
                conversion_type: conversionType,
                value: value,
                timestamp: Date.now()
            });
        }
    }

    // SEO and Content Performance
    trackContentEngagement() {
        const contentElements = document.querySelectorAll('article, .content, .post-content');
        
        contentElements.forEach((element, index) => {
            const observer = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        this.sendAnalyticsEvent('content_view', {
                            content_index: index,
                            content_id: element.id || null,
                            content_type: element.dataset.type || 'unknown',
                            viewport_percentage: Math.round(entry.intersectionRatio * 100),
                            timestamp: Date.now()
                        });
                    }
                });
            }, {
                threshold: [0.25, 0.5, 0.75, 1.0]
            });

            observer.observe(element);
        });
    }

    // Conversion Tracking
    trackConversion(type, value = null, metadata = {}) {
        const conversionData = {
            event_type: 'conversion',
            conversion_type: type,
            value: value,
            session_id: this.sessionId,
            visitor_id: this.visitorId,
            timestamp: Date.now(),
            ...metadata
        };

        this.sendAnalyticsEvent('conversion', conversionData);
    }

    // Error Tracking
    trackError(error, context = {}) {
        const errorData = {
            event_type: 'javascript_error',
            error_message: error.message,
            error_stack: error.stack,
            error_line: error.lineno,
            error_column: error.colno,
            error_filename: error.filename,
            page_url: window.location.href,
            user_agent: navigator.userAgent,
            timestamp: Date.now(),
            ...context
        };

        this.sendAnalyticsEvent('error', errorData, true);
    }

    // Utility Methods
    startSessionTimer() {
        this.sessionStartTime = Date.now();
    }

    generateSessionId() {
        return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    getOrCreateVisitorId() {
        let visitorId = localStorage.getItem('visitor_id');
        
        if (!visitorId) {
            visitorId = 'visitor_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
            localStorage.setItem('visitor_id', visitorId);
        }
        
        return visitorId;
    }

    getButtonType(element) {
        if (element.classList.contains('btn-primary')) return 'primary';
        if (element.classList.contains('btn-secondary')) return 'secondary';
        if (element.classList.contains('btn-outline')) return 'outline';
        return 'unknown';
    }

    async sendAnalyticsEvent(eventType, data, immediate = false) {
        const eventData = {
            event_type: eventType,
            session_id: this.sessionId,
            visitor_id: this.visitorId,
            url: window.location.href,
            timestamp: Date.now(),
            ...data
        };

        // Store locally
        this.storeEvent(eventData);

        // In a real implementation, send to analytics API
        if (immediate) {
            // Use sendBeacon for immediate events (page unload, etc.)
            if (navigator.sendBeacon) {
                navigator.sendBeacon('/api/v1/analytics', JSON.stringify(eventData));
            }
        } else {
            // Queue for batch sending
            this.queueEvent(eventData);
        }
    }

    storeEvent(event) {
        try {
            const events = JSON.parse(localStorage.getItem('analytics_events') || '[]');
            events.push(event);
            
            // Keep only last 200 events
            if (events.length > 200) {
                events.splice(0, events.length - 200);
            }
            
            localStorage.setItem('analytics_events', JSON.stringify(events));
        } catch (error) {
            console.error('Failed to store analytics event:', error);
        }
    }

    queueEvent(event) {
        // In a real implementation, this would queue events for batch sending
        console.log('Analytics event:', event);
    }
}

// Global error handler
window.addEventListener('error', (event) => {
    if (window.autonomousAnalytics) {
        window.autonomousAnalytics.trackError(event.error || event, {
            type: 'window_error'
        });
    }
});

// Global unhandled promise rejection handler
window.addEventListener('unhandledrejection', (event) => {
    if (window.autonomousAnalytics) {
        window.autonomousAnalytics.trackError(new Error(event.reason), {
            type: 'unhandled_promise_rejection'
        });
    }
});

// Initialize analytics when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.autonomousAnalytics = new AutonomousAnalytics();
});