-- Database Schema for Autonomous Content Creation Service

-- Create necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enums for entity statuses

-- Client status enum
CREATE TYPE client_status AS ENUM (
    'Active',
    'Inactive',
    'Pending',
    'Suspended'
);

-- Project status enum
CREATE TYPE project_status AS ENUM (
    'Draft',
    'Planning',
    'InProgress',
    'Review',
    'Completed',
    'Cancelled'
);

-- Priority enum
CREATE TYPE priority AS ENUM (
    'High',
    'Medium',
    'Low'
);

-- Content type enum
CREATE TYPE content_type AS ENUM (
    'BlogPost',
    'SocialPost',
    'EmailNewsletter',
    'WebsiteCopy',
    'TechnicalArticle',
    'ProductDescription',
    'PressRelease'
);

-- Content status enum
CREATE TYPE content_status AS ENUM (
    'Planning',
    'Researching',
    'Drafting',
    'Editing',
    'Review',
    'Approved',
    'Published',
    'Archived'
);

-- Transaction type enum
CREATE TYPE transaction_type AS ENUM (
    'Payment',
    'Refund',
    'Fee',
    'Subscription',
    'ApiCost'
);

-- Transaction status enum
CREATE TYPE transaction_status AS ENUM (
    'Pending',
    'Completed',
    'Failed',
    'Disputed'
);

-- Payment method enum
CREATE TYPE payment_method_type AS ENUM (
    'CreditCard',
    'BankTransfer',
    'Cryptocurrency',
    'PayPal'
);

-- Feedback source enum
CREATE TYPE feedback_source AS ENUM (
    'Client',
    'System',
    'ThirdParty'
);

-- Feedback type enum
CREATE TYPE feedback_type AS ENUM (
    'Revision',
    'Comment',
    'Rating'
);

-- Feedback status enum
CREATE TYPE feedback_status AS ENUM (
    'New',
    'Acknowledged',
    'Implemented',
    'Rejected'
);

-- Capability type enum
CREATE TYPE capability_type AS ENUM (
    'ContentCreation',
    'ContentEditing',
    'ClientCommunication',
    'MarketAnalysis',
    'FinancialTransaction'
);

-- Capability status enum
CREATE TYPE capability_status AS ENUM (
    'Active',
    'Learning',
    'Deprecated'
);

-- Clients table
CREATE TABLE clients (
    client_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255) NOT NULL UNIQUE,
    contact_phone VARCHAR(50) NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    status client_status NOT NULL DEFAULT 'Pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    billing_street VARCHAR(255) NOT NULL,
    billing_city VARCHAR(255) NOT NULL,
    billing_state VARCHAR(255),
    billing_postal_code VARCHAR(50),
    billing_country VARCHAR(255) NOT NULL
);

-- Create index on client status
CREATE INDEX idx_clients_status ON clients(status);
-- Create index on client email
CREATE INDEX idx_clients_email ON clients(contact_email);

-- Client profiles table
CREATE TABLE client_profiles (
    profile_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(client_id) ON DELETE CASCADE,
    industry VARCHAR(255) NOT NULL,
    brand_voice TEXT,
    target_audience TEXT,
    content_goals JSONB NOT NULL,
    style_preferences JSONB,
    example_content TEXT[],
    competitor_urls TEXT[],
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create unique index on client_id to ensure one profile per client
CREATE UNIQUE INDEX idx_client_profiles_client_id ON client_profiles(client_id);

-- Projects table
CREATE TABLE projects (
    project_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(client_id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    content_type content_type NOT NULL,
    deadline TIMESTAMP NOT NULL,
    budget_amount DECIMAL(10, 2) NOT NULL,
    budget_currency VARCHAR(3) NOT NULL,
    priority priority NOT NULL DEFAULT 'Medium',
    status project_status NOT NULL DEFAULT 'Draft',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on client_id
CREATE INDEX idx_projects_client_id ON projects(client_id);
-- Create index on project status
CREATE INDEX idx_projects_status ON projects(status);
-- Create index on deadline for efficient deadline queries
CREATE INDEX idx_projects_deadline ON projects(deadline);

-- Content table
CREATE TABLE content (
    content_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    type content_type NOT NULL,
    status content_status NOT NULL DEFAULT 'Planning',
    data TEXT,
    metadata JSONB,
    version INT NOT NULL DEFAULT 1,
    word_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    readability_score DECIMAL(5, 2),
    seo_score DECIMAL(5, 2),
    engagement_score DECIMAL(5, 2),
    plagiarism_score DECIMAL(5, 2)
);

-- Create index on project_id
CREATE INDEX idx_content_project_id ON content(project_id);
-- Create index on content status
CREATE INDEX idx_content_status ON content(status);
-- Create index on content type
CREATE INDEX idx_content_type ON content(type);

-- Content versions table
CREATE TABLE content_versions (
    version_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_id UUID NOT NULL REFERENCES content(content_id) ON DELETE CASCADE,
    version_number INT NOT NULL,
    data TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL
);

-- Create index on content_id
CREATE INDEX idx_content_versions_content_id ON content_versions(content_id);
-- Create unique index on content_id and version_number
CREATE UNIQUE INDEX idx_content_versions_unique ON content_versions(content_id, version_number);

-- Transactions table
CREATE TABLE transactions (
    transaction_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(client_id),
    project_id UUID REFERENCES projects(project_id),
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    type transaction_type NOT NULL,
    status transaction_status NOT NULL DEFAULT 'Pending',
    payment_method payment_method_type NOT NULL,
    external_reference VARCHAR(255),
    description TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on client_id
CREATE INDEX idx_transactions_client_id ON transactions(client_id);
-- Create index on project_id
CREATE INDEX idx_transactions_project_id ON transactions(project_id);
-- Create index on transaction status
CREATE INDEX idx_transactions_status ON transactions(status);
-- Create index on transaction type
CREATE INDEX idx_transactions_type ON transactions(type);
-- Create index on timestamp for date range queries
CREATE INDEX idx_transactions_timestamp ON transactions(timestamp);

-- Feedback table
CREATE TABLE feedback (
    feedback_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_id UUID REFERENCES content(content_id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(project_id) ON DELETE CASCADE,
    source feedback_source NOT NULL,
    type feedback_type NOT NULL,
    score DECIMAL(3, 1),
    comment TEXT,
    status feedback_status NOT NULL DEFAULT 'New',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (content_id IS NOT NULL OR project_id IS NOT NULL)
);

-- Create index on content_id
CREATE INDEX idx_feedback_content_id ON feedback(content_id);
-- Create index on project_id
CREATE INDEX idx_feedback_project_id ON feedback(project_id);
-- Create index on feedback status
CREATE INDEX idx_feedback_status ON feedback(status);
-- Create index on feedback source
CREATE INDEX idx_feedback_source ON feedback(source);

-- System capabilities table
CREATE TABLE system_capabilities (
    capability_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    type capability_type NOT NULL,
    status capability_status NOT NULL DEFAULT 'Learning',
    performance_metrics JSONB,
    api_dependencies TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on capability type
CREATE INDEX idx_system_capabilities_type ON system_capabilities(type);
-- Create index on capability status
CREATE INDEX idx_system_capabilities_status ON system_capabilities(status);

-- Events table for event sourcing
CREATE TABLE events (
    event_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(255) NOT NULL,
    aggregate_id UUID NOT NULL,
    occurred_at TIMESTAMP NOT NULL,
    event_data JSONB NOT NULL
);

-- Create index on event_type
CREATE INDEX idx_events_type ON events(event_type);
-- Create index on aggregate_id
CREATE INDEX idx_events_aggregate_id ON events(aggregate_id);
-- Create index on occurred_at for time-based queries
CREATE INDEX idx_events_occurred_at ON events(occurred_at);
