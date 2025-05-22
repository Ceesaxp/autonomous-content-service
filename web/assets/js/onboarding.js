class OnboardingWidget {
    constructor() {
        this.apiBaseUrl = '/api/v1';
        this.currentSession = null;
        this.currentStage = 'initial';
        this.responses = {};
        this.isLoading = false;
        
        this.init();
    }
    
    init() {
        this.createOnboardingModal();
        this.bindEvents();
        this.loadProgressSave();
    }
    
    createOnboardingModal() {
        const modal = document.createElement('div');
        modal.id = 'onboarding-modal';
        modal.className = 'onboarding-modal';
        modal.innerHTML = `
            <div class="onboarding-content">
                <div class="onboarding-header">
                    <div class="progress-container">
                        <div class="progress-bar">
                            <div class="progress-fill" id="progress-fill"></div>
                        </div>
                        <span class="progress-text" id="progress-text">Step 1 of 8</span>
                    </div>
                    <button class="close-btn" id="close-onboarding">&times;</button>
                </div>
                
                <div class="onboarding-body">
                    <div class="stage-indicator" id="stage-indicator">
                        <h2 id="stage-title">Welcome</h2>
                        <p id="stage-description">Let's get your content strategy set up!</p>
                    </div>
                    
                    <div class="conversation-area" id="conversation-area">
                        <div class="messages" id="messages-container"></div>
                        <div class="input-area" id="input-area">
                            <div class="message-input-container">
                                <textarea 
                                    id="message-input" 
                                    placeholder="Type your response here..."
                                    rows="3"
                                ></textarea>
                                <button id="send-message" class="send-btn">
                                    <span class="send-icon">→</span>
                                </button>
                            </div>
                        </div>
                    </div>
                    
                    <div class="questions-area" id="questions-area">
                        <!-- Dynamic questions will be inserted here -->
                    </div>
                </div>
                
                <div class="onboarding-footer">
                    <button id="prev-step" class="btn btn-secondary" style="display: none;">Previous</button>
                    <button id="next-step" class="btn btn-primary" style="display: none;">Next</button>
                    <div class="loading-indicator" id="loading-indicator" style="display: none;">
                        <div class="spinner"></div>
                        <span>Processing...</span>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
    }
    
    bindEvents() {
        // Modal controls
        document.getElementById('close-onboarding').addEventListener('click', () => {
            this.closeModal();
        });
        
        // Message sending
        document.getElementById('send-message').addEventListener('click', () => {
            this.sendMessage();
        });
        
        document.getElementById('message-input').addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });
        
        // Navigation
        document.getElementById('prev-step').addEventListener('click', () => {
            this.previousStep();
        });
        
        document.getElementById('next-step').addEventListener('click', () => {
            this.nextStep();
        });
        
        // Auto-resize textarea
        const messageInput = document.getElementById('message-input');
        messageInput.addEventListener('input', () => {
            messageInput.style.height = 'auto';
            messageInput.style.height = Math.min(messageInput.scrollHeight, 120) + 'px';
        });
    }
    
    async startOnboarding(clientId = null) {
        try {
            this.showLoading(true);
            
            const response = await fetch(`${this.apiBaseUrl}/onboarding/start`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    clientId: clientId || this.generateTempClientId()
                })
            });
            
            if (!response.ok) {
                throw new Error('Failed to start onboarding');
            }
            
            const session = await response.json();
            this.currentSession = session;
            this.saveProgress();
            
            this.showModal();
            this.displayInitialMessage();
            this.updateProgress();
            
        } catch (error) {
            console.error('Failed to start onboarding:', error);
            this.showError('Failed to start onboarding. Please try again.');
        } finally {
            this.showLoading(false);
        }
    }
    
    async sendMessage() {
        const messageInput = document.getElementById('message-input');
        const message = messageInput.value.trim();
        
        if (!message || this.isLoading) return;
        
        try {
            this.showLoading(true);
            this.addMessage('client', message);
            messageInput.value = '';
            messageInput.style.height = 'auto';
            
            const response = await fetch(`${this.apiBaseUrl}/onboarding/message`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    sessionId: this.currentSession.sessionId,
                    message: message
                })
            });
            
            if (!response.ok) {
                throw new Error('Failed to send message');
            }
            
            const result = await response.json();
            this.handleResponse(result);
            this.updateProgress();
            this.saveProgress();
            
        } catch (error) {
            console.error('Failed to send message:', error);
            this.showError('Failed to send message. Please try again.');
        } finally {
            this.showLoading(false);
        }
    }
    
    handleResponse(response) {
        // Add system message
        this.addMessage('system', response.message);
        
        // Update current stage
        this.currentStage = response.stage;
        this.updateStageInfo(response.stage);
        
        // Display questions if any
        if (response.questions && response.questions.length > 0) {
            this.displayQuestions(response.questions);
        }
        
        // Handle completion
        if (response.stage === 'complete') {
            this.handleOnboardingComplete(response);
        }
        
        // Update UI state
        this.updateNavigationButtons(response);
    }
    
    displayQuestions(questions) {
        const questionsArea = document.getElementById('questions-area');
        questionsArea.innerHTML = '';
        
        questions.forEach((question, index) => {
            const questionElement = this.createQuestionElement(question, index);
            questionsArea.appendChild(questionElement);
        });
        
        questionsArea.style.display = 'block';
    }
    
    createQuestionElement(question, index) {
        const questionDiv = document.createElement('div');
        questionDiv.className = 'question-item';
        questionDiv.dataset.questionId = question.id;
        
        let inputHtml = '';
        
        switch (question.type) {
            case 'text':
                inputHtml = `
                    <textarea 
                        id="q-${question.id}" 
                        class="question-input"
                        placeholder="${question.description || 'Enter your response...'}"
                        ${question.required ? 'required' : ''}
                    ></textarea>
                `;
                break;
                
            case 'choice':
                inputHtml = `
                    <div class="choice-options">
                        ${question.options.map(option => `
                            <label class="choice-option">
                                <input 
                                    type="radio" 
                                    name="q-${question.id}" 
                                    value="${option.value}"
                                    ${question.required ? 'required' : ''}
                                >
                                <span class="choice-label">${option.label}</span>
                                ${option.description ? `<small class="choice-description">${option.description}</small>` : ''}
                            </label>
                        `).join('')}
                    </div>
                `;
                break;
                
            case 'multiple':
                inputHtml = `
                    <div class="multiple-options">
                        ${question.options.map(option => `
                            <label class="multiple-option">
                                <input 
                                    type="checkbox" 
                                    name="q-${question.id}" 
                                    value="${option.value}"
                                >
                                <span class="multiple-label">${option.label}</span>
                                ${option.description ? `<small class="multiple-description">${option.description}</small>` : ''}
                            </label>
                        `).join('')}
                    </div>
                `;
                break;
                
            case 'scale':
                const min = question.validation?.minValue || 1;
                const max = question.validation?.maxValue || 10;
                inputHtml = `
                    <div class="scale-input">
                        <input 
                            type="range" 
                            id="q-${question.id}" 
                            min="${min}" 
                            max="${max}" 
                            value="${Math.round((min + max) / 2)}"
                            class="scale-slider"
                        >
                        <div class="scale-labels">
                            <span>${min}</span>
                            <span id="scale-value-${question.id}">${Math.round((min + max) / 2)}</span>
                            <span>${max}</span>
                        </div>
                    </div>
                `;
                break;
        }
        
        questionDiv.innerHTML = `
            <div class="question-header">
                <h4 class="question-title">${question.question}</h4>
                ${question.description ? `<p class="question-description">${question.description}</p>` : ''}
                ${question.required ? '<span class="required-indicator">*</span>' : ''}
            </div>
            <div class="question-input-area">
                ${inputHtml}
            </div>
        `;
        
        // Add event listeners for scale inputs
        if (question.type === 'scale') {
            setTimeout(() => {
                const slider = questionDiv.querySelector(`#q-${question.id}`);
                const valueDisplay = questionDiv.querySelector(`#scale-value-${question.id}`);
                
                slider.addEventListener('input', () => {
                    valueDisplay.textContent = slider.value;
                });
            }, 0);
        }
        
        return questionDiv;
    }
    
    addMessage(speaker, message) {
        const messagesContainer = document.getElementById('messages-container');
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${speaker}-message`;
        
        messageDiv.innerHTML = `
            <div class="message-content">
                <div class="message-text">${this.formatMessage(message)}</div>
                <div class="message-time">${new Date().toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}</div>
            </div>
        `;
        
        messagesContainer.appendChild(messageDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
    
    formatMessage(message) {
        // Simple formatting - in production, use a proper markdown parser
        return message
            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
            .replace(/\*(.*?)\*/g, '<em>$1</em>')
            .replace(/\n/g, '<br>');
    }
    
    updateProgress() {
        const stages = [
            'initial', 'industry', 'goals', 'audience', 
            'style', 'brand', 'competitors', 'welcome', 'complete'
        ];
        
        const currentIndex = stages.indexOf(this.currentStage);
        const progress = (currentIndex / (stages.length - 1)) * 100;
        
        const progressFill = document.getElementById('progress-fill');
        const progressText = document.getElementById('progress-text');
        
        progressFill.style.width = `${progress}%`;
        progressText.textContent = `Step ${currentIndex + 1} of ${stages.length}`;
    }
    
    updateStageInfo(stage) {
        const stageTitle = document.getElementById('stage-title');
        const stageDescription = document.getElementById('stage-description');
        
        const stageInfo = {
            'initial': {
                title: 'Welcome',
                description: "Let's get your content strategy set up!"
            },
            'industry': {
                title: 'About Your Business',
                description: 'Tell us about your industry and company'
            },
            'goals': {
                title: 'Your Goals',
                description: 'What do you want to achieve with content?'
            },
            'audience': {
                title: 'Target Audience',
                description: 'Who are you trying to reach?'
            },
            'style': {
                title: 'Content Style',
                description: 'Define your preferred tone and style'
            },
            'brand': {
                title: 'Brand Voice',
                description: 'Capture your unique brand personality'
            },
            'competitors': {
                title: 'Competitive Landscape',
                description: 'Understand your market position'
            },
            'welcome': {
                title: 'Almost Done!',
                description: 'Just a few final details'
            },
            'complete': {
                title: 'Complete!',
                description: 'Your content strategy is ready'
            }
        };
        
        const info = stageInfo[stage] || stageInfo['initial'];
        stageTitle.textContent = info.title;
        stageDescription.textContent = info.description;
    }
    
    updateNavigationButtons(response) {
        const prevBtn = document.getElementById('prev-step');
        const nextBtn = document.getElementById('next-step');
        
        // Show/hide navigation based on stage
        if (this.currentStage === 'initial') {
            prevBtn.style.display = 'none';
        } else {
            prevBtn.style.display = 'inline-block';
        }
        
        if (this.currentStage === 'complete') {
            nextBtn.style.display = 'none';
        } else if (response.questions && response.questions.length > 0) {
            nextBtn.style.display = 'inline-block';
            nextBtn.textContent = this.currentStage === 'welcome' ? 'Complete' : 'Next';
        }
    }
    
    displayInitialMessage() {
        this.addMessage('system', "Hi! I'm here to help you create a personalized content strategy. This process takes about 10-15 minutes and will help me understand your business, goals, and preferences. Ready to get started?");
    }
    
    showModal() {
        const modal = document.getElementById('onboarding-modal');
        modal.style.display = 'flex';
        document.body.classList.add('modal-open');
    }
    
    closeModal() {
        const modal = document.getElementById('onboarding-modal');
        modal.style.display = 'none';
        document.body.classList.remove('modal-open');
        this.saveProgress();
    }
    
    showLoading(show) {
        this.isLoading = show;
        const loadingIndicator = document.getElementById('loading-indicator');
        const sendBtn = document.getElementById('send-message');
        const messageInput = document.getElementById('message-input');
        
        if (show) {
            loadingIndicator.style.display = 'flex';
            sendBtn.disabled = true;
            messageInput.disabled = true;
        } else {
            loadingIndicator.style.display = 'none';
            sendBtn.disabled = false;
            messageInput.disabled = false;
        }
    }
    
    showError(message) {
        this.addMessage('system', `⚠️ ${message}`);
    }
    
    handleOnboardingComplete(response) {
        // Clear questions area
        document.getElementById('questions-area').style.display = 'none';
        
        // Show completion message
        this.addMessage('system', "🎉 Congratulations! Your onboarding is complete. I'll now create your personalized content strategy and client profile.");
        
        // Show next steps
        setTimeout(() => {
            const nextSteps = response.metadata?.next_steps || [
                'Review your content strategy',
                'Start your first project',
                'Access your client dashboard'
            ];
            
            const stepsHtml = nextSteps.map(step => `• ${step}`).join('<br>');
            this.addMessage('system', `Here's what you can do next:<br><br>${stepsHtml}`);
            
            // Show completion button
            const nextBtn = document.getElementById('next-step');
            nextBtn.textContent = 'Access Dashboard';
            nextBtn.style.display = 'inline-block';
            nextBtn.onclick = () => {
                this.redirectToDashboard();
            };
        }, 2000);
        
        // Clear saved progress
        this.clearProgress();
    }
    
    redirectToDashboard() {
        // In a real implementation, redirect to the client dashboard
        window.location.href = '/dashboard';
    }
    
    generateTempClientId() {
        // Generate a temporary client ID for guest users
        return 'temp_' + Math.random().toString(36).substr(2, 9);
    }
    
    saveProgress() {
        if (this.currentSession) {
            localStorage.setItem('onboarding_progress', JSON.stringify({
                sessionId: this.currentSession.sessionId,
                stage: this.currentStage,
                responses: this.responses,
                timestamp: Date.now()
            }));
        }
    }
    
    loadProgressSave() {
        const saved = localStorage.getItem('onboarding_progress');
        if (saved) {
            try {
                const data = JSON.parse(saved);
                const age = Date.now() - data.timestamp;
                
                // Only restore if less than 24 hours old
                if (age < 24 * 60 * 60 * 1000) {
                    this.currentSession = { sessionId: data.sessionId };
                    this.currentStage = data.stage;
                    this.responses = data.responses;
                    
                    // Show resume option
                    this.showResumeOption();
                }
            } catch (error) {
                console.error('Failed to load saved progress:', error);
            }
        }
    }
    
    showResumeOption() {
        // Create a small notification to resume onboarding
        const notification = document.createElement('div');
        notification.className = 'resume-notification';
        notification.innerHTML = `
            <div class="resume-content">
                <span>Continue your onboarding where you left off?</span>
                <button id="resume-onboarding" class="btn btn-sm btn-primary">Resume</button>
                <button id="start-new-onboarding" class="btn btn-sm btn-secondary">Start New</button>
            </div>
        `;
        
        document.body.appendChild(notification);
        
        document.getElementById('resume-onboarding').addEventListener('click', () => {
            this.resumeOnboarding();
            notification.remove();
        });
        
        document.getElementById('start-new-onboarding').addEventListener('click', () => {
            this.clearProgress();
            this.startOnboarding();
            notification.remove();
        });
        
        // Auto-hide after 10 seconds
        setTimeout(() => {
            if (notification.parentNode) {
                notification.remove();
            }
        }, 10000);
    }
    
    async resumeOnboarding() {
        try {
            this.showLoading(true);
            this.showModal();
            
            // Get current session state
            const response = await fetch(`${this.apiBaseUrl}/onboarding/session/${this.currentSession.sessionId}`);
            if (response.ok) {
                const session = await response.json();
                this.currentSession = session;
                
                // Restore conversation
                this.restoreConversation(session.conversationLog);
                this.updateStageInfo(this.currentStage);
                this.updateProgress();
                
                // Get current questions
                const questionsResponse = await fetch(`${this.apiBaseUrl}/onboarding/questions/${this.currentSession.sessionId}`);
                if (questionsResponse.ok) {
                    const questions = await questionsResponse.json();
                    if (questions.length > 0) {
                        this.displayQuestions(questions);
                    }
                }
            }
        } catch (error) {
            console.error('Failed to resume onboarding:', error);
            this.showError('Failed to resume onboarding. Starting fresh.');
            this.startOnboarding();
        } finally {
            this.showLoading(false);
        }
    }
    
    restoreConversation(conversationLog) {
        const messagesContainer = document.getElementById('messages-container');
        messagesContainer.innerHTML = '';
        
        conversationLog.forEach(msg => {
            this.addMessage(msg.speaker, msg.message);
        });
    }
    
    clearProgress() {
        localStorage.removeItem('onboarding_progress');
        this.currentSession = null;
        this.currentStage = 'initial';
        this.responses = {};
    }
    
    // Public API methods
    static create() {
        return new OnboardingWidget();
    }
    
    // Method to trigger onboarding from external buttons
    static startOnboardingFlow(clientId = null) {
        const widget = new OnboardingWidget();
        widget.startOnboarding(clientId);
        return widget;
    }
}

// Auto-initialize if not already done
if (typeof window !== 'undefined') {
    window.OnboardingWidget = OnboardingWidget;
    
    // Auto-start onboarding if there's a trigger
    document.addEventListener('DOMContentLoaded', () => {
        // Check for onboarding triggers
        const startButtons = document.querySelectorAll('[data-start-onboarding]');
        startButtons.forEach(button => {
            button.addEventListener('click', (e) => {
                e.preventDefault();
                const clientId = button.dataset.clientId || null;
                OnboardingWidget.startOnboardingFlow(clientId);
            });
        });
        
        // Auto-start if URL parameter exists
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.get('onboarding') === 'start') {
            OnboardingWidget.startOnboardingFlow();
        }
    });
}