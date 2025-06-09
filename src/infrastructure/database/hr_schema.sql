-- HR Management Database Schema for Autonomous Content Creation Service

-- HR-specific enums

CREATE TYPE talent_type AS ENUM (
    'Human',
    'AI'
);

CREATE TYPE talent_status AS ENUM (
    'Available',
    'Engaged',
    'Unavailable',
    'Offboarded',
    'Deprecated'
);

CREATE TYPE engagement_status AS ENUM (
    'Draft',
    'Pending',
    'Active',
    'Paused',
    'Completed',
    'Terminated'
);

CREATE TYPE engagement_type AS ENUM (
    'FullTime',
    'PartTime',
    'Contract',
    'Project',
    'APIAccess'
);

CREATE TYPE skill_level AS ENUM (
    'Beginner',
    'Intermediate',
    'Advanced',
    'Expert'
);

CREATE TYPE training_status AS ENUM (
    'NotStarted',
    'InProgress',
    'Completed',
    'Failed',
    'Expired'
);

CREATE TYPE performance_rating AS ENUM (
    'Exceptional',
    'ExceedsExpectations',
    'MeetsExpectations',
    'NeedsImprovement',
    'Unsatisfactory'
);

CREATE TYPE application_status AS ENUM (
    'New',
    'Screening',
    'Interview',
    'Assessment',
    'Approved',
    'Rejected',
    'Withdrawn'
);

-- Core HR Tables

-- Talent table (base for both human and AI)
CREATE TABLE talent (
    talent_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type talent_type NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    status talent_status NOT NULL DEFAULT 'Available',
    reputation_score DECIMAL(3, 2) DEFAULT 0.0,
    currency VARCHAR(3) DEFAULT 'USD',
    hourly_rate_amount DECIMAL(12, 2),
    location VARCHAR(255),
    timezone VARCHAR(50),
    availability JSONB DEFAULT '{}',
    profile_data JSONB DEFAULT '{}',
    last_active_at TIMESTAMP,
    onboarded_at TIMESTAMP,
    offboarded_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_reputation_score CHECK (reputation_score >= 0 AND reputation_score <= 5.0),
    CONSTRAINT email_required_for_humans CHECK (type = 'AI' OR email IS NOT NULL)
);

-- Indexes for talent table
CREATE INDEX idx_talent_type ON talent(type);
CREATE INDEX idx_talent_status ON talent(status);
CREATE INDEX idx_talent_email ON talent(email);
CREATE INDEX idx_talent_reputation ON talent(reputation_score DESC);
CREATE INDEX idx_talent_location ON talent(location);
CREATE INDEX idx_talent_created_at ON talent(created_at);

-- Human-specific additional fields
CREATE TABLE human_contributors (
    talent_id UUID PRIMARY KEY REFERENCES talent(talent_id) ON DELETE CASCADE,
    phone VARCHAR(50),
    linkedin_url VARCHAR(500),
    portfolio_url VARCHAR(500),
    years_experience INTEGER DEFAULT 0,
    preferred_hours INTEGER DEFAULT 40,
    languages TEXT[] DEFAULT '{}',
    work_authorization JSONB DEFAULT '{}',
    
    CONSTRAINT positive_experience CHECK (years_experience >= 0),
    CONSTRAINT valid_preferred_hours CHECK (preferred_hours > 0 AND preferred_hours <= 168)
);

-- AI agent-specific additional fields
CREATE TABLE ai_agents (
    talent_id UUID PRIMARY KEY REFERENCES talent(talent_id) ON DELETE CASCADE,
    provider VARCHAR(255) NOT NULL,
    model VARCHAR(255) NOT NULL,
    api_endpoint VARCHAR(500) NOT NULL,
    api_version VARCHAR(50),
    capabilities TEXT[] DEFAULT '{}',
    rate_limits JSONB DEFAULT '{}',
    cost_per_request_amount DECIMAL(10, 6),
    cost_per_request_currency VARCHAR(3) DEFAULT 'USD',
    cost_per_token_amount DECIMAL(10, 8),
    cost_per_token_currency VARCHAR(3) DEFAULT 'USD',
    max_tokens INTEGER,
    response_time_ms INTEGER,
    reliability DECIMAL(3, 2) DEFAULT 1.0,
    last_health_check TIMESTAMP,
    
    CONSTRAINT valid_reliability CHECK (reliability >= 0 AND reliability <= 1.0),
    CONSTRAINT positive_response_time CHECK (response_time_ms >= 0)
);

-- Skills table
CREATE TABLE skills (
    skill_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    level skill_level NOT NULL,
    years_used DECIMAL(4, 1) DEFAULT 0,
    last_used TIMESTAMP,
    verified BOOLEAN DEFAULT FALSE,
    verified_by VARCHAR(255),
    verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT positive_years_used CHECK (years_used >= 0),
    UNIQUE(talent_id, name)
);

CREATE INDEX idx_skills_talent_id ON skills(talent_id);
CREATE INDEX idx_skills_name ON skills(name);
CREATE INDEX idx_skills_category ON skills(category);
CREATE INDEX idx_skills_level ON skills(level);

-- Certifications table
CREATE TABLE certifications (
    certification_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    issuer VARCHAR(255) NOT NULL,
    credential_id VARCHAR(255),
    issue_date DATE NOT NULL,
    expiry_date DATE,
    verification_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_date_range CHECK (expiry_date IS NULL OR expiry_date > issue_date)
);

CREATE INDEX idx_certifications_talent_id ON certifications(talent_id);
CREATE INDEX idx_certifications_expiry ON certifications(expiry_date);
CREATE INDEX idx_certifications_issuer ON certifications(issuer);

-- Engagements table
CREATE TABLE engagements (
    engagement_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    type engagement_type NOT NULL,
    status engagement_status NOT NULL DEFAULT 'Draft',
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE,
    hours_per_week INTEGER DEFAULT 40,
    rate_type VARCHAR(50) DEFAULT 'Hourly', -- Hourly, Fixed, Retainer
    rate_amount DECIMAL(12, 2),
    currency VARCHAR(3) DEFAULT 'USD',
    contract_id UUID,
    manager_id UUID,
    team_id UUID,
    performance_metrics JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_date_range CHECK (end_date IS NULL OR end_date >= start_date),
    CONSTRAINT positive_hours CHECK (hours_per_week > 0 AND hours_per_week <= 168),
    CONSTRAINT positive_rate CHECK (rate_amount IS NULL OR rate_amount >= 0)
);

CREATE INDEX idx_engagements_talent_id ON engagements(talent_id);
CREATE INDEX idx_engagements_status ON engagements(status);
CREATE INDEX idx_engagements_type ON engagements(type);
CREATE INDEX idx_engagements_start_date ON engagements(start_date);
CREATE INDEX idx_engagements_end_date ON engagements(end_date);

-- Work assignments table
CREATE TABLE work_assignments (
    assignment_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    engagement_id UUID NOT NULL REFERENCES engagements(engagement_id) ON DELETE CASCADE,
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(project_id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority priority NOT NULL DEFAULT 'Medium',
    status VARCHAR(50) DEFAULT 'Created',
    estimated_hours DECIMAL(6, 2) DEFAULT 0,
    actual_hours DECIMAL(6, 2) DEFAULT 0,
    due_date TIMESTAMP,
    completed_at TIMESTAMP,
    quality_score DECIMAL(3, 2),
    feedback_notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT positive_hours CHECK (estimated_hours >= 0 AND actual_hours >= 0),
    CONSTRAINT valid_quality_score CHECK (quality_score IS NULL OR (quality_score >= 0 AND quality_score <= 5.0))
);

CREATE INDEX idx_assignments_engagement_id ON work_assignments(engagement_id);
CREATE INDEX idx_assignments_talent_id ON work_assignments(talent_id);
CREATE INDEX idx_assignments_project_id ON work_assignments(project_id);
CREATE INDEX idx_assignments_status ON work_assignments(status);
CREATE INDEX idx_assignments_due_date ON work_assignments(due_date);
CREATE INDEX idx_assignments_priority ON work_assignments(priority);

-- Deliverables table
CREATE TABLE deliverables (
    deliverable_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    assignment_id UUID NOT NULL REFERENCES work_assignments(assignment_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(100) DEFAULT 'Document',
    status VARCHAR(50) DEFAULT 'Pending',
    file_url VARCHAR(500),
    metadata JSONB DEFAULT '{}',
    submitted_at TIMESTAMP,
    accepted_at TIMESTAMP,
    rejected_at TIMESTAMP,
    rejection_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deliverables_assignment_id ON deliverables(assignment_id);
CREATE INDEX idx_deliverables_status ON deliverables(status);
CREATE INDEX idx_deliverables_type ON deliverables(type);

-- Performance reviews table
CREATE TABLE performance_reviews (
    review_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    engagement_id UUID REFERENCES engagements(engagement_id),
    reviewer_id UUID,
    review_period_start DATE NOT NULL,
    review_period_end DATE NOT NULL,
    overall_rating performance_rating NOT NULL,
    quality_score DECIMAL(3, 2) DEFAULT 0,
    productivity_score DECIMAL(3, 2) DEFAULT 0,
    reliability_score DECIMAL(3, 2) DEFAULT 0,
    communication_score DECIMAL(3, 2) DEFAULT 0,
    strengths TEXT[],
    improvement_areas TEXT[],
    goals TEXT[],
    comments TEXT,
    metrics JSONB DEFAULT '{}',
    compensation_adjustment_amount DECIMAL(12, 2),
    compensation_adjustment_currency VARCHAR(3),
    next_review_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_review_period CHECK (review_period_end >= review_period_start),
    CONSTRAINT valid_scores CHECK (
        quality_score >= 0 AND quality_score <= 5 AND
        productivity_score >= 0 AND productivity_score <= 5 AND
        reliability_score >= 0 AND reliability_score <= 5 AND
        communication_score >= 0 AND communication_score <= 5
    )
);

CREATE INDEX idx_performance_reviews_talent_id ON performance_reviews(talent_id);
CREATE INDEX idx_performance_reviews_engagement_id ON performance_reviews(engagement_id);
CREATE INDEX idx_performance_reviews_period ON performance_reviews(review_period_start, review_period_end);
CREATE INDEX idx_performance_reviews_rating ON performance_reviews(overall_rating);

-- Compensation plans table
CREATE TABLE compensation_plans (
    compensation_plan_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    engagement_id UUID REFERENCES engagements(engagement_id),
    type VARCHAR(50) NOT NULL DEFAULT 'Hourly', -- Salary, Hourly, Project, Retainer
    base_amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_frequency VARCHAR(50) DEFAULT 'Weekly', -- Weekly, BiWeekly, Monthly
    bonus_structure JSONB DEFAULT '{}',
    benefits TEXT[],
    effective_date DATE NOT NULL,
    end_date DATE,
    tax_withholding DECIMAL(5, 4) DEFAULT 0,
    payment_method_id UUID,
    smart_contract_addr VARCHAR(42),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT positive_base_amount CHECK (base_amount > 0),
    CONSTRAINT valid_date_range CHECK (end_date IS NULL OR end_date >= effective_date),
    CONSTRAINT valid_tax_withholding CHECK (tax_withholding >= 0 AND tax_withholding <= 1)
);

CREATE INDEX idx_compensation_plans_talent_id ON compensation_plans(talent_id);
CREATE INDEX idx_compensation_plans_engagement_id ON compensation_plans(engagement_id);
CREATE INDEX idx_compensation_plans_effective_date ON compensation_plans(effective_date);
CREATE INDEX idx_compensation_plans_type ON compensation_plans(type);

-- Payroll records table
CREATE TABLE payroll_records (
    payroll_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    engagement_id UUID REFERENCES engagements(engagement_id),
    pay_period_start DATE NOT NULL,
    pay_period_end DATE NOT NULL,
    gross_amount DECIMAL(12, 2) NOT NULL,
    net_amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    hours_worked DECIMAL(6, 2) DEFAULT 0,
    deductions JSONB DEFAULT '{}',
    bonuses JSONB DEFAULT '{}',
    payment_date DATE NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    transaction_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'Pending',
    tax_document_urls TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_pay_period CHECK (pay_period_end >= pay_period_start),
    CONSTRAINT positive_amounts CHECK (gross_amount >= 0 AND net_amount >= 0),
    CONSTRAINT positive_hours CHECK (hours_worked >= 0)
);

CREATE INDEX idx_payroll_records_talent_id ON payroll_records(talent_id);
CREATE INDEX idx_payroll_records_engagement_id ON payroll_records(engagement_id);
CREATE INDEX idx_payroll_records_pay_period ON payroll_records(pay_period_start, pay_period_end);
CREATE INDEX idx_payroll_records_payment_date ON payroll_records(payment_date);
CREATE INDEX idx_payroll_records_status ON payroll_records(status);

-- Training programs table
CREATE TABLE training_programs (
    training_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(100) DEFAULT 'Skill', -- Onboarding, Skill, Compliance
    target_audience VARCHAR(100),
    duration INTEGER DEFAULT 0, -- in hours
    format VARCHAR(50) DEFAULT 'Online', -- Online, InPerson, Hybrid
    prerequisites TEXT[],
    learning_objectives TEXT[],
    passing_score DECIMAL(3, 2) DEFAULT 70.0,
    certification_id UUID,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT positive_duration CHECK (duration >= 0),
    CONSTRAINT valid_passing_score CHECK (passing_score >= 0 AND passing_score <= 100)
);

CREATE INDEX idx_training_programs_type ON training_programs(type);
CREATE INDEX idx_training_programs_target_audience ON training_programs(target_audience);
CREATE INDEX idx_training_programs_active ON training_programs(is_active);

-- Training materials table
CREATE TABLE training_materials (
    material_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    training_id UUID NOT NULL REFERENCES training_programs(training_id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    type VARCHAR(50) DEFAULT 'Document', -- Video, Document, Quiz, Exercise
    content_url VARCHAR(500),
    duration INTEGER DEFAULT 0, -- in minutes
    order_index INTEGER DEFAULT 0,
    is_required BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT positive_duration CHECK (duration >= 0),
    CONSTRAINT positive_order CHECK (order_index >= 0)
);

CREATE INDEX idx_training_materials_training_id ON training_materials(training_id);
CREATE INDEX idx_training_materials_type ON training_materials(type);
CREATE INDEX idx_training_materials_order ON training_materials(training_id, order_index);

-- Training progress table
CREATE TABLE training_progress (
    progress_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    training_id UUID NOT NULL REFERENCES training_programs(training_id) ON DELETE CASCADE,
    status training_status NOT NULL DEFAULT 'NotStarted',
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    progress DECIMAL(5, 2) DEFAULT 0, -- 0-100
    score DECIMAL(5, 2),
    attempts INTEGER DEFAULT 0,
    material_progress JSONB DEFAULT '{}',
    certificate_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_progress CHECK (progress >= 0 AND progress <= 100),
    CONSTRAINT valid_score CHECK (score IS NULL OR (score >= 0 AND score <= 100)),
    CONSTRAINT positive_attempts CHECK (attempts >= 0),
    UNIQUE(talent_id, training_id)
);

CREATE INDEX idx_training_progress_talent_id ON training_progress(talent_id);
CREATE INDEX idx_training_progress_training_id ON training_progress(training_id);
CREATE INDEX idx_training_progress_status ON training_progress(status);

-- Job postings table
CREATE TABLE job_postings (
    job_posting_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    type engagement_type NOT NULL,
    department VARCHAR(100),
    required_skills TEXT[],
    preferred_skills TEXT[],
    experience_years INTEGER DEFAULT 0,
    education_level VARCHAR(100),
    location VARCHAR(255),
    remote BOOLEAN DEFAULT FALSE,
    salary_range JSONB DEFAULT '{}',
    benefits TEXT[],
    posted_date DATE NOT NULL DEFAULT CURRENT_DATE,
    closing_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    application_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT positive_experience CHECK (experience_years >= 0),
    CONSTRAINT valid_closing_date CHECK (closing_date IS NULL OR closing_date >= posted_date),
    CONSTRAINT positive_application_count CHECK (application_count >= 0)
);

CREATE INDEX idx_job_postings_type ON job_postings(type);
CREATE INDEX idx_job_postings_department ON job_postings(department);
CREATE INDEX idx_job_postings_location ON job_postings(location);
CREATE INDEX idx_job_postings_remote ON job_postings(remote);
CREATE INDEX idx_job_postings_active ON job_postings(is_active);
CREATE INDEX idx_job_postings_posted_date ON job_postings(posted_date);

-- Talent applications table
CREATE TABLE talent_applications (
    application_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    job_posting_id UUID NOT NULL REFERENCES job_postings(job_posting_id) ON DELETE CASCADE,
    status application_status NOT NULL DEFAULT 'New',
    cover_letter TEXT,
    resume_url VARCHAR(500),
    portfolio_urls TEXT[],
    screening_score DECIMAL(3, 2),
    screening_notes TEXT,
    interview_notes TEXT,
    assessment_score DECIMAL(3, 2),
    reference_checks JSONB DEFAULT '{}',
    decision_date DATE,
    decision_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_screening_score CHECK (screening_score IS NULL OR (screening_score >= 0 AND screening_score <= 5)),
    CONSTRAINT valid_assessment_score CHECK (assessment_score IS NULL OR (assessment_score >= 0 AND assessment_score <= 5)),
    UNIQUE(talent_id, job_posting_id)
);

CREATE INDEX idx_talent_applications_talent_id ON talent_applications(talent_id);
CREATE INDEX idx_talent_applications_job_posting_id ON talent_applications(job_posting_id);
CREATE INDEX idx_talent_applications_status ON talent_applications(status);
CREATE INDEX idx_talent_applications_created_at ON talent_applications(created_at);

-- Contractor agreements table
CREATE TABLE contractor_agreements (
    agreement_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    engagement_id UUID REFERENCES engagements(engagement_id),
    contract_type VARCHAR(100) NOT NULL,
    template_id UUID NOT NULL,
    terms JSONB DEFAULT '{}',
    start_date DATE NOT NULL,
    end_date DATE,
    renewal_date DATE,
    signed_at TIMESTAMP,
    signature_id VARCHAR(255),
    document_url VARCHAR(500) NOT NULL,
    status VARCHAR(50) DEFAULT 'Draft',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_date_range CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_contractor_agreements_talent_id ON contractor_agreements(talent_id);
CREATE INDEX idx_contractor_agreements_engagement_id ON contractor_agreements(engagement_id);
CREATE INDEX idx_contractor_agreements_status ON contractor_agreements(status);
CREATE INDEX idx_contractor_agreements_start_date ON contractor_agreements(start_date);
CREATE INDEX idx_contractor_agreements_end_date ON contractor_agreements(end_date);

-- Compliance checks table
CREATE TABLE compliance_checks (
    check_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    check_type VARCHAR(100) NOT NULL,
    provider VARCHAR(100),
    status VARCHAR(50) DEFAULT 'Pending',
    result VARCHAR(50),
    details JSONB DEFAULT '{}',
    document_urls TEXT[],
    valid_until DATE,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_compliance_checks_talent_id ON compliance_checks(talent_id);
CREATE INDEX idx_compliance_checks_type ON compliance_checks(check_type);
CREATE INDEX idx_compliance_checks_status ON compliance_checks(status);
CREATE INDEX idx_compliance_checks_valid_until ON compliance_checks(valid_until);

-- Offboarding checklists table
CREATE TABLE offboarding_checklists (
    checklist_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talent_id UUID NOT NULL REFERENCES talent(talent_id) ON DELETE CASCADE,
    engagement_id UUID REFERENCES engagements(engagement_id),
    reason VARCHAR(255),
    last_working_date DATE NOT NULL,
    knowledge_transfer JSONB DEFAULT '{}',
    exit_interview_url VARCHAR(500),
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_offboarding_checklists_talent_id ON offboarding_checklists(talent_id);
CREATE INDEX idx_offboarding_checklists_engagement_id ON offboarding_checklists(engagement_id);
CREATE INDEX idx_offboarding_checklists_last_working_date ON offboarding_checklists(last_working_date);

-- Offboarding tasks table
CREATE TABLE offboarding_tasks (
    task_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    checklist_id UUID NOT NULL REFERENCES offboarding_checklists(checklist_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    assigned_to UUID,
    due_date DATE NOT NULL,
    completed_at TIMESTAMP,
    completed_by UUID,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_offboarding_tasks_checklist_id ON offboarding_tasks(checklist_id);
CREATE INDEX idx_offboarding_tasks_assigned_to ON offboarding_tasks(assigned_to);
CREATE INDEX idx_offboarding_tasks_due_date ON offboarding_tasks(due_date);
CREATE INDEX idx_offboarding_tasks_completed_at ON offboarding_tasks(completed_at);

-- Triggers for automatic timestamp updates

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_talent_updated_at BEFORE UPDATE ON talent
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_engagements_updated_at BEFORE UPDATE ON engagements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_work_assignments_updated_at BEFORE UPDATE ON work_assignments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_deliverables_updated_at BEFORE UPDATE ON deliverables
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_compensation_plans_updated_at BEFORE UPDATE ON compensation_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_training_programs_updated_at BEFORE UPDATE ON training_programs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_training_progress_updated_at BEFORE UPDATE ON training_progress
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_job_postings_updated_at BEFORE UPDATE ON job_postings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_talent_applications_updated_at BEFORE UPDATE ON talent_applications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_contractor_agreements_updated_at BEFORE UPDATE ON contractor_agreements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_offboarding_checklists_updated_at BEFORE UPDATE ON offboarding_checklists
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Views for common queries

-- Active talent view
CREATE VIEW active_talent AS
SELECT 
    t.*,
    array_agg(DISTINCT s.name) AS skill_names,
    array_agg(DISTINCT c.name) AS certification_names
FROM talent t
LEFT JOIN skills s ON t.talent_id = s.talent_id
LEFT JOIN certifications c ON t.talent_id = c.talent_id AND c.is_active = true
WHERE t.status IN ('Available', 'Engaged')
GROUP BY t.talent_id;

-- Active engagements view
CREATE VIEW active_engagements AS
SELECT 
    e.*,
    t.name AS talent_name,
    t.type AS talent_type,
    COUNT(wa.assignment_id) AS assignment_count,
    AVG(wa.quality_score) AS avg_quality_score
FROM engagements e
JOIN talent t ON e.talent_id = t.talent_id
LEFT JOIN work_assignments wa ON e.engagement_id = wa.engagement_id
WHERE e.status = 'Active'
GROUP BY e.engagement_id, t.talent_id;

-- Performance summary view
CREATE VIEW talent_performance_summary AS
SELECT 
    t.talent_id,
    t.name,
    t.type,
    COUNT(DISTINCT e.engagement_id) AS total_engagements,
    COUNT(DISTINCT wa.assignment_id) AS total_assignments,
    AVG(wa.quality_score) AS avg_quality_score,
    AVG(pr.overall_rating::text::numeric) AS avg_performance_rating,
    COUNT(DISTINCT pr.review_id) AS total_reviews,
    SUM(wa.actual_hours) AS total_hours_worked
FROM talent t
LEFT JOIN engagements e ON t.talent_id = e.talent_id
LEFT JOIN work_assignments wa ON e.engagement_id = wa.engagement_id
LEFT JOIN performance_reviews pr ON t.talent_id = pr.talent_id
GROUP BY t.talent_id;

-- Training completion view
CREATE VIEW training_completion_summary AS
SELECT 
    t.talent_id,
    t.name,
    COUNT(tp.training_id) AS total_trainings,
    COUNT(CASE WHEN tp.status = 'Completed' THEN 1 END) AS completed_trainings,
    COUNT(CASE WHEN tp.status = 'InProgress' THEN 1 END) AS in_progress_trainings,
    AVG(CASE WHEN tp.status = 'Completed' THEN tp.score END) AS avg_training_score
FROM talent t
LEFT JOIN training_progress tp ON t.talent_id = tp.talent_id
GROUP BY t.talent_id;