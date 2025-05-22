// Chat functionality for autonomous customer service

class ChatWidget {
    constructor() {
        this.isOpen = false;
        this.messages = [];
        this.isTyping = false;
        this.apiEndpoint = '/api/v1/chat'; // Will connect to Go backend
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.addWelcomeMessage();
    }

    setupEventListeners() {
        const chatToggle = document.querySelector('.chat-toggle');
        const chatWindow = document.getElementById('chat-window');
        const chatInput = document.getElementById('chat-input');
        
        if (chatToggle) {
            chatToggle.addEventListener('click', () => this.toggleChat());
        }
        
        if (chatInput) {
            chatInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.sendMessage();
                }
            });
        }
    }

    toggleChat() {
        const chatWindow = document.getElementById('chat-window');
        this.isOpen = !this.isOpen;
        
        if (chatWindow) {
            chatWindow.style.display = this.isOpen ? 'flex' : 'none';
            
            if (this.isOpen) {
                this.focusInput();
                this.trackChatOpen();
            }
        }
    }

    addWelcomeMessage() {
        this.addMessage({
            text: "Hello! I'm the autonomous assistant for our content creation service. I can help you with:",
            type: 'bot',
            options: [
                "Get a price quote",
                "Learn about our services", 
                "View portfolio examples",
                "Ask about turnaround times",
                "Speak to a human"
            ]
        });
    }

    async sendMessage() {
        const chatInput = document.getElementById('chat-input');
        const message = chatInput.value.trim();
        
        if (!message) return;
        
        // Add user message
        this.addMessage({
            text: message,
            type: 'user'
        });
        
        // Clear input
        chatInput.value = '';
        
        // Show typing indicator
        this.showTyping();
        
        try {
            // Get bot response
            const response = await this.getBotResponse(message);
            this.hideTyping();
            this.addMessage(response);
            
        } catch (error) {
            this.hideTyping();
            this.addMessage({
                text: "I apologize, but I'm having trouble responding right now. Please try again or contact us directly at hello@autonomous-content.com",
                type: 'bot'
            });
            console.error('Chat error:', error);
        }
    }

    addMessage(messageData) {
        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;
        
        const messageElement = document.createElement('div');
        messageElement.className = `message ${messageData.type}-message`;
        
        // Add message text
        const textElement = document.createElement('div');
        textElement.textContent = messageData.text;
        messageElement.appendChild(textElement);
        
        // Add options if provided
        if (messageData.options && messageData.options.length > 0) {
            const optionsContainer = document.createElement('div');
            optionsContainer.className = 'chat-options';
            optionsContainer.style.marginTop = '10px';
            
            messageData.options.forEach(option => {
                const optionButton = document.createElement('button');
                optionButton.textContent = option;
                optionButton.className = 'chat-option';
                optionButton.style.cssText = `
                    display: block;
                    width: 100%;
                    margin: 5px 0;
                    padding: 8px 12px;
                    background: var(--bg-secondary);
                    border: 1px solid var(--border-light);
                    border-radius: var(--radius-sm);
                    cursor: pointer;
                    text-align: left;
                    font-size: 0.875rem;
                `;
                
                optionButton.addEventListener('click', () => {
                    this.handleOptionClick(option);
                });
                
                optionButton.addEventListener('mouseenter', () => {
                    optionButton.style.background = 'var(--primary-color)';
                    optionButton.style.color = 'white';
                });
                
                optionButton.addEventListener('mouseleave', () => {
                    optionButton.style.background = 'var(--bg-secondary)';
                    optionButton.style.color = 'var(--text-primary)';
                });
                
                optionsContainer.appendChild(optionButton);
            });
            
            messageElement.appendChild(optionsContainer);
        }
        
        messagesContainer.appendChild(messageElement);
        
        // Scroll to bottom
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
        
        // Store message
        this.messages.push(messageData);
    }

    handleOptionClick(option) {
        // Simulate user clicking on an option
        const chatInput = document.getElementById('chat-input');
        if (chatInput) {
            chatInput.value = option;
            this.sendMessage();
        }
    }

    showTyping() {
        if (this.isTyping) return;
        
        this.isTyping = true;
        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;
        
        const typingElement = document.createElement('div');
        typingElement.className = 'message bot-message typing-indicator';
        typingElement.innerHTML = `
            <div class="typing-dots">
                <span></span>
                <span></span>
                <span></span>
            </div>
        `;
        
        // Add CSS for typing animation
        const style = document.createElement('style');
        style.textContent = `
            .typing-dots {
                display: flex;
                gap: 4px;
                align-items: center;
            }
            .typing-dots span {
                width: 6px;
                height: 6px;
                background: var(--text-secondary);
                border-radius: 50%;
                animation: typing 1.4s infinite ease-in-out;
            }
            .typing-dots span:nth-child(1) { animation-delay: -0.32s; }
            .typing-dots span:nth-child(2) { animation-delay: -0.16s; }
            @keyframes typing {
                0%, 80%, 100% { opacity: 0.3; transform: scale(0.8); }
                40% { opacity: 1; transform: scale(1); }
            }
        `;
        document.head.appendChild(style);
        
        messagesContainer.appendChild(typingElement);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    hideTyping() {
        this.isTyping = false;
        const typingIndicator = document.querySelector('.typing-indicator');
        if (typingIndicator) {
            typingIndicator.remove();
        }
    }

    async getBotResponse(userMessage) {
        // In a real implementation, this would call the Go backend API
        // For now, provide intelligent responses based on keywords
        
        const lowerMessage = userMessage.toLowerCase();
        
        // Simulate API delay
        await this.delay(1000 + Math.random() * 2000);
        
        // Price-related queries
        if (lowerMessage.includes('price') || lowerMessage.includes('cost') || lowerMessage.includes('quote')) {
            return {
                text: "I'd be happy to help you get a price quote! Our pricing depends on several factors:",
                type: 'bot',
                options: [
                    "Blog posts start at $50",
                    "Marketing copy starts at $75", 
                    "Technical docs start at $100",
                    "Get detailed quote"
                ]
            };
        }
        
        // Service-related queries  
        if (lowerMessage.includes('service') || lowerMessage.includes('what do you do')) {
            return {
                text: "We offer autonomous content creation services including:",
                type: 'bot',
                options: [
                    "Blog Posts & Articles",
                    "Marketing Copy",
                    "Technical Documentation", 
                    "Social Media Content",
                    "View all services"
                ]
            };
        }
        
        // Timeline queries
        if (lowerMessage.includes('time') || lowerMessage.includes('fast') || lowerMessage.includes('delivery')) {
            return {
                text: "Our autonomous system is incredibly fast! Standard delivery is 24-48 hours, but we offer rush options:",
                type: 'bot',
                options: [
                    "Rush delivery: 6-12 hours",
                    "Express delivery: 12-24 hours", 
                    "Standard delivery: 24-48 hours",
                    "Start my project now"
                ]
            };
        }
        
        // Portfolio queries
        if (lowerMessage.includes('example') || lowerMessage.includes('portfolio') || lowerMessage.includes('sample')) {
            return {
                text: "I'd love to show you our work! We have examples across many industries:",
                type: 'bot',
                options: [
                    "View blog post examples",
                    "See marketing copy samples",
                    "Technical documentation examples",
                    "Browse full portfolio"
                ]
            };
        }
        
        // Human handoff
        if (lowerMessage.includes('human') || lowerMessage.includes('person') || lowerMessage.includes('speak to someone')) {
            return {
                text: "While our system operates autonomously, I can connect you with our team for complex questions. Here are your options:",
                type: 'bot',
                options: [
                    "Schedule a call",
                    "Send an email to hello@autonomous-content.com",
                    "Continue with me for now",
                    "Submit a project request"
                ]
            };
        }
        
        // Greeting responses
        if (lowerMessage.includes('hello') || lowerMessage.includes('hi') || lowerMessage.includes('hey')) {
            return {
                text: "Hello! Great to meet you. I'm here to help you with your content needs. What would you like to know?",
                type: 'bot',
                options: [
                    "Get a price quote",
                    "Learn about our services",
                    "See example work",
                    "Ask about delivery times"
                ]
            };
        }
        
        // Default response with helpful options
        return {
            text: "I understand you're interested in our content creation services. Let me help you find what you need:",
            type: 'bot',
            options: [
                "Get pricing information",
                "Learn about our services", 
                "View portfolio examples",
                "Start a project",
                "Speak to our team"
            ]
        };
    }

    focusInput() {
        const chatInput = document.getElementById('chat-input');
        if (chatInput) {
            setTimeout(() => chatInput.focus(), 100);
        }
    }

    trackChatOpen() {
        if (window.websiteController && window.websiteController.analytics) {
            window.websiteController.analytics.trackEvent('chat_opened', {
                page: window.location.pathname,
                timestamp: new Date().toISOString()
            });
        }
    }

    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}

// Global functions for HTML handlers
window.toggleChat = function() {
    if (window.chatWidget) {
        window.chatWidget.toggleChat();
    }
};

window.sendChatMessage = function() {
    if (window.chatWidget) {
        window.chatWidget.sendMessage();
    }
};

window.handleChatKeypress = function(event) {
    if (event.key === 'Enter') {
        window.sendChatMessage();
    }
};

// Initialize chat widget when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.chatWidget = new ChatWidget();
});