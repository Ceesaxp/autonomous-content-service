-- Legal and Compliance Database Schema

-- Enum types for legal entities
CREATE TYPE contract_type AS ENUM ('Service', 'Employment', 'NDA', 'License', 'Partnership', 'Vendor', 'Client');
CREATE TYPE contract_status AS ENUM ('Draft', 'Review', 'Pending', 'Signed', 'Active', 'Expired', 'Terminated', 'Disputed', 'Archived');
CREATE TYPE term_type AS ENUM ('Payment', 'Delivery', 'IP', 'Confidentiality', 'Termination', 'Liability', 'Dispute', 'Governing');
CREATE TYPE signature_type AS ENUM ('Electronic', 'Digital', 'Wet', 'DocuSign');
CREATE TYPE signature_status AS ENUM ('Pending', 'Signed', 'Declined', 'Expired', 'Invalid');
CREATE TYPE compliance_type AS ENUM ('GDPR', 'CCPA', 'SOX', 'HIPAA', 'SOC2', 'ISO27001', 'COPPA');
CREATE TYPE compliance_status AS ENUM ('Compliant', 'NonCompliant', 'Pending', 'Exempt', 'Unknown');
CREATE TYPE license_type AS ENUM ('Exclusive', 'NonExclusive', 'Sole', 'CreativeCommons', 'MIT', 'GPL', 'Proprietary');
CREATE TYPE ip_type AS ENUM ('Copyright', 'Trademark', 'Patent', 'TradeSecret', 'Software', 'Content');
CREATE TYPE usage_right AS ENUM ('Use', 'Modify', 'Distribute', 'Sublicense', 'Commercial', 'Attribution');
CREATE TYPE insurance_type AS ENUM ('General', 'Professional', 'Cyber', 'Errors', 'Directors');
CREATE TYPE policy_status AS ENUM ('Active', 'Inactive', 'Expired', 'Cancelled', 'Pending');
CREATE TYPE dispute_type AS ENUM ('Breach', 'Payment', 'Delivery', 'Quality', 'IP', 'Termination');
CREATE TYPE dispute_status AS ENUM ('Open', 'Mediation', 'Arbitration', 'Litigation', 'Resolved', 'Closed');
CREATE TYPE resolution_method AS ENUM ('Negotiation', 'Mediation', 'Arbitration', 'Litigation');
CREATE TYPE risk_level AS ENUM ('Low', 'Medium', 'High', 'Critical');
CREATE TYPE report_type AS ENUM ('Quarterly', 'Annual', 'Monthly', 'Adhoc', 'Incident');
CREATE TYPE report_status AS ENUM ('Draft', 'Review', 'Pending', 'Filed', 'Rejected', 'Amended');

-- Contract templates table
CREATE TABLE contract_templates (
    template_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type contract_type NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    content TEXT NOT NULL,
    parameters JSONB,
    default_terms JSONB,
    metadata JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(name, version)
);

-- Template parameters table
CREATE TABLE template_parameters (
    param_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID REFERENCES contract_templates(template_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    param_type VARCHAR(50) NOT NULL,
    description TEXT,
    required BOOLEAN DEFAULT false,
    default_value JSONB,
    validation TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(template_id, name)
);

-- Contracts table
CREATE TABLE contracts (
    contract_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    type contract_type NOT NULL,
    status contract_status DEFAULT 'Draft',
    version INTEGER DEFAULT 1,
    parent_contract_id UUID REFERENCES contracts(contract_id),
    client_id UUID NOT NULL,
    project_id UUID,
    template_id UUID REFERENCES contract_templates(template_id),
    content TEXT NOT NULL,
    parameters JSONB,
    terms JSONB,
    signatures JSONB,
    effective_date TIMESTAMP WITH TIME ZONE,
    expiration_date TIMESTAMP WITH TIME ZONE,
    renewal_date TIMESTAMP WITH TIME ZONE,
    compliance_checks JSONB,
    dispute_resolution JSONB,
    ip_licenses JSONB,
    insurance_required BOOLEAN DEFAULT false,
    insurance_policies JSONB,
    risk_assessment JSONB,
    audit_trail JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    archived_at TIMESTAMP WITH TIME ZONE,
    
    INDEX idx_contracts_client_id (client_id),
    INDEX idx_contracts_project_id (project_id),
    INDEX idx_contracts_status (status),
    INDEX idx_contracts_type (type),
    INDEX idx_contracts_expiration (expiration_date),
    INDEX idx_contracts_created_at (created_at)
);

-- Contract terms table
CREATE TABLE contract_terms (
    term_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID REFERENCES contracts(contract_id) ON DELETE CASCADE,
    type term_type NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    content TEXT NOT NULL,
    is_mandatory BOOLEAN DEFAULT false,
    order_index INTEGER,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_contract_terms_contract_id (contract_id),
    INDEX idx_contract_terms_type (type)
);

-- Contract signatures table
CREATE TABLE contract_signatures (
    signature_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID REFERENCES contracts(contract_id) ON DELETE CASCADE,
    signer_name VARCHAR(255) NOT NULL,
    signer_email VARCHAR(255) NOT NULL,
    signer_role VARCHAR(100),
    signature_type signature_type NOT NULL,
    signature_data TEXT,
    signature_hash VARCHAR(512),
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status signature_status DEFAULT 'Pending',
    verification_hash VARCHAR(512),
    certificate_id VARCHAR(255),
    is_valid BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_signatures_contract_id (contract_id),
    INDEX idx_signatures_signer_email (signer_email),
    INDEX idx_signatures_status (status),
    INDEX idx_signatures_timestamp (timestamp)
);

-- Compliance checks table
CREATE TABLE compliance_checks (
    check_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50), -- 'contract', 'content', 'system', etc.
    entity_id UUID,
    type compliance_type NOT NULL,
    regulation VARCHAR(100) NOT NULL,
    requirement TEXT,
    status compliance_status DEFAULT 'Pending',
    result TEXT,
    evidence JSONB,
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    checked_by VARCHAR(100),
    next_check TIMESTAMP WITH TIME ZONE,
    remediation TEXT,
    risk_level risk_level DEFAULT 'Low',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_compliance_checks_entity (entity_type, entity_id),
    INDEX idx_compliance_checks_type (type),
    INDEX idx_compliance_checks_regulation (regulation),
    INDEX idx_compliance_checks_status (status),
    INDEX idx_compliance_checks_next_check (next_check)
);

-- IP licenses table
CREATE TABLE ip_licenses (
    license_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type license_type NOT NULL,
    name VARCHAR(255) NOT NULL,
    licensor_name VARCHAR(255) NOT NULL,
    licensee_name VARCHAR(255) NOT NULL,
    ip_type ip_type NOT NULL,
    ip_description TEXT,
    usage_rights TEXT[], -- Array of usage_right enum values
    restrictions TEXT[],
    territory VARCHAR(255) DEFAULT 'Worldwide',
    effective_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expiration_date TIMESTAMP WITH TIME ZONE,
    royalty_rate DECIMAL(10, 4),
    fee_amount DECIMAL(15, 2),
    fee_currency CHAR(3),
    is_exclusive BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_ip_licenses_type (type),
    INDEX idx_ip_licenses_ip_type (ip_type),
    INDEX idx_ip_licenses_licensor (licensor_name),
    INDEX idx_ip_licenses_licensee (licensee_name),
    INDEX idx_ip_licenses_expiration (expiration_date),
    INDEX idx_ip_licenses_active (is_active)
);

-- IP usage tracking table
CREATE TABLE ip_usage_events (
    event_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    license_id UUID REFERENCES ip_licenses(license_id) ON DELETE CASCADE,
    content_id UUID,
    project_id UUID,
    usage_type VARCHAR(50), -- 'creation', 'modification', 'distribution', etc.
    usage_details JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_id VARCHAR(100),
    metadata JSONB,
    
    INDEX idx_ip_usage_license_id (license_id),
    INDEX idx_ip_usage_content_id (content_id),
    INDEX idx_ip_usage_timestamp (timestamp)
);

-- Insurance policies table
CREATE TABLE insurance_policies (
    policy_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type insurance_type NOT NULL,
    policy_number VARCHAR(100) UNIQUE NOT NULL,
    provider VARCHAR(255) NOT NULL,
    coverage_amount DECIMAL(15, 2),
    coverage_currency CHAR(3),
    deductible_amount DECIMAL(15, 2),
    deductible_currency CHAR(3),
    premium_amount DECIMAL(15, 2),
    premium_currency CHAR(3),
    effective_date TIMESTAMP WITH TIME ZONE NOT NULL,
    expiration_date TIMESTAMP WITH TIME ZONE NOT NULL,
    renewal_date TIMESTAMP WITH TIME ZONE,
    coverage_details TEXT[],
    exclusions TEXT[],
    status policy_status DEFAULT 'Active',
    is_active BOOLEAN DEFAULT true,
    document_url TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_insurance_policies_type (type),
    INDEX idx_insurance_policies_provider (provider),
    INDEX idx_insurance_policies_status (status),
    INDEX idx_insurance_policies_expiration (expiration_date),
    INDEX idx_insurance_policies_renewal (renewal_date)
);

-- Dispute resolution table
CREATE TABLE dispute_resolutions (
    dispute_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID REFERENCES contracts(contract_id) ON DELETE CASCADE,
    type dispute_type NOT NULL,
    status dispute_status DEFAULT 'Open',
    description TEXT NOT NULL,
    initiated_by VARCHAR(255) NOT NULL,
    resolution_method resolution_method DEFAULT 'Negotiation',
    mediator VARCHAR(255),
    arbitrator VARCHAR(255),
    venue VARCHAR(255),
    governing_law VARCHAR(100),
    timeline JSONB,
    resolution TEXT,
    cost_amount DECIMAL(15, 2),
    cost_currency CHAR(3),
    initiated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_dispute_resolutions_contract_id (contract_id),
    INDEX idx_dispute_resolutions_type (type),
    INDEX idx_dispute_resolutions_status (status),
    INDEX idx_dispute_resolutions_initiated_at (initiated_at)
);

-- Dispute events table
CREATE TABLE dispute_events (
    event_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dispute_id UUID REFERENCES dispute_resolutions(dispute_id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    actor VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    evidence TEXT[],
    metadata JSONB,
    
    INDEX idx_dispute_events_dispute_id (dispute_id),
    INDEX idx_dispute_events_timestamp (timestamp)
);

-- Legal risk assessments table
CREATE TABLE legal_risk_assessments (
    assessment_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50), -- 'contract', 'project', 'system', etc.
    entity_id UUID,
    risk_level risk_level NOT NULL,
    risk_score DECIMAL(5, 3) CHECK (risk_score >= 0 AND risk_score <= 1),
    risk_factors JSONB,
    recommendations TEXT[],
    required_clauses TEXT[],
    compliance_issues TEXT[],
    insurance_required BOOLEAN DEFAULT false,
    legal_review BOOLEAN DEFAULT false,
    assessed_by VARCHAR(100) NOT NULL,
    assessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    
    INDEX idx_risk_assessments_entity (entity_type, entity_id),
    INDEX idx_risk_assessments_risk_level (risk_level),
    INDEX idx_risk_assessments_assessed_at (assessed_at),
    INDEX idx_risk_assessments_expires_at (expires_at)
);

-- Regulatory reports table
CREATE TABLE regulatory_reports (
    report_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type report_type NOT NULL,
    regulation VARCHAR(100) NOT NULL,
    authority VARCHAR(255) NOT NULL,
    period VARCHAR(50), -- 'Q1-2024', '2024', 'Jan-2024', etc.
    status report_status DEFAULT 'Draft',
    content TEXT,
    data JSONB,
    filing_deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    filed_at TIMESTAMP WITH TIME ZONE,
    confirmation_id VARCHAR(255),
    document_url TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_regulatory_reports_type (type),
    INDEX idx_regulatory_reports_regulation (regulation),
    INDEX idx_regulatory_reports_authority (authority),
    INDEX idx_regulatory_reports_status (status),
    INDEX idx_regulatory_reports_filing_deadline (filing_deadline),
    INDEX idx_regulatory_reports_period (period)
);

-- Contract audit trail table
CREATE TABLE contract_audit_entries (
    entry_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID REFERENCES contracts(contract_id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    field VARCHAR(100),
    old_value TEXT,
    new_value TEXT,
    user_id VARCHAR(100),
    user_agent TEXT,
    ip_address INET,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    hash VARCHAR(512) NOT NULL,
    metadata JSONB,
    
    INDEX idx_audit_entries_contract_id (contract_id),
    INDEX idx_audit_entries_timestamp (timestamp),
    INDEX idx_audit_entries_action (action),
    INDEX idx_audit_entries_user_id (user_id)
);

-- Legal notifications and alerts table
CREATE TABLE legal_alerts (
    alert_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL, -- 'contract_expiring', 'compliance_due', 'signature_pending', etc.
    severity risk_level DEFAULT 'Medium',
    title VARCHAR(255) NOT NULL,
    description TEXT,
    entity_type VARCHAR(50),
    entity_id UUID,
    triggered_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    due_date TIMESTAMP WITH TIME ZONE,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by VARCHAR(100),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB,
    
    INDEX idx_legal_alerts_type (type),
    INDEX idx_legal_alerts_severity (severity),
    INDEX idx_legal_alerts_entity (entity_type, entity_id),
    INDEX idx_legal_alerts_triggered_at (triggered_at),
    INDEX idx_legal_alerts_due_date (due_date),
    INDEX idx_legal_alerts_active (is_active)
);

-- Compliance requirements table (for storing regulatory requirements)
CREATE TABLE compliance_requirements (
    requirement_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    regulation VARCHAR(100) NOT NULL,
    requirement_code VARCHAR(50),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(100),
    applicable_to VARCHAR(100), -- 'contracts', 'data_processing', 'financial', etc.
    enforcement_level VARCHAR(50) DEFAULT 'mandatory', -- 'mandatory', 'recommended', 'optional'
    penalty_description TEXT,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    effective_date TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    is_active BOOLEAN DEFAULT true,
    
    INDEX idx_compliance_requirements_regulation (regulation),
    INDEX idx_compliance_requirements_category (category),
    INDEX idx_compliance_requirements_applicable_to (applicable_to),
    INDEX idx_compliance_requirements_active (is_active),
    UNIQUE(regulation, requirement_code)
);

-- Data processing records table (for GDPR compliance)
CREATE TABLE data_processing_records (
    record_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50), -- 'client', 'project', 'content', etc.
    entity_id UUID,
    processing_purpose VARCHAR(255) NOT NULL,
    legal_basis VARCHAR(100) NOT NULL, -- 'consent', 'contract', 'legitimate_interest', etc.
    data_categories TEXT[], -- 'personal_data', 'sensitive_data', etc.
    data_subjects VARCHAR(255), -- 'clients', 'employees', 'website_visitors', etc.
    retention_period INTERVAL,
    security_measures TEXT[],
    third_party_transfers JSONB,
    consent_obtained BOOLEAN DEFAULT false,
    consent_date TIMESTAMP WITH TIME ZONE,
    consent_withdrawn BOOLEAN DEFAULT false,
    consent_withdrawn_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_data_processing_entity (entity_type, entity_id),
    INDEX idx_data_processing_purpose (processing_purpose),
    INDEX idx_data_processing_legal_basis (legal_basis),
    INDEX idx_data_processing_consent (consent_obtained)
);

-- Create indexes for performance optimization
CREATE INDEX idx_contracts_full_text ON contracts USING gin(to_tsvector('english', title || ' ' || content));
CREATE INDEX idx_compliance_checks_due ON compliance_checks (next_check) WHERE next_check IS NOT NULL AND status != 'Compliant';
CREATE INDEX idx_contracts_renewal_due ON contracts (renewal_date) WHERE renewal_date IS NOT NULL AND status = 'Active';
CREATE INDEX idx_insurance_renewal_due ON insurance_policies (renewal_date) WHERE renewal_date IS NOT NULL AND status = 'Active';
CREATE INDEX idx_signatures_pending ON contract_signatures (contract_id) WHERE status = 'Pending';

-- Triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_contracts_updated_at BEFORE UPDATE ON contracts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_contract_templates_updated_at BEFORE UPDATE ON contract_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ip_licenses_updated_at BEFORE UPDATE ON ip_licenses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_insurance_policies_updated_at BEFORE UPDATE ON insurance_policies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_dispute_resolutions_updated_at BEFORE UPDATE ON dispute_resolutions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_regulatory_reports_updated_at BEFORE UPDATE ON regulatory_reports FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_data_processing_records_updated_at BEFORE UPDATE ON data_processing_records FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert some default compliance requirements
INSERT INTO compliance_requirements (regulation, requirement_code, title, description, category, applicable_to) VALUES
('GDPR', 'ART6', 'Lawfulness of Processing', 'Processing shall be lawful only if and to the extent that at least one legal basis applies', 'data_protection', 'data_processing'),
('GDPR', 'ART7', 'Conditions for Consent', 'Where processing is based on consent, the controller shall demonstrate that consent has been given', 'data_protection', 'data_processing'),
('GDPR', 'ART17', 'Right to Erasure', 'The data subject shall have the right to obtain erasure of personal data', 'data_subject_rights', 'data_processing'),
('GDPR', 'ART20', 'Right to Data Portability', 'The data subject shall have the right to receive personal data in a structured format', 'data_subject_rights', 'data_processing'),
('CCPA', 'SEC1798.100', 'Right to Know', 'Consumers have the right to know what personal information is collected', 'consumer_rights', 'data_processing'),
('CCPA', 'SEC1798.105', 'Right to Delete', 'Consumers have the right to request deletion of personal information', 'consumer_rights', 'data_processing'),
('SOX', 'SEC302', 'Corporate Responsibility', 'Principal executive and financial officers must certify financial reports', 'financial_reporting', 'financial'),
('SOX', 'SEC404', 'Internal Control Assessment', 'Management must assess effectiveness of internal control over financial reporting', 'financial_reporting', 'financial');

-- Insert default contract templates
INSERT INTO contract_templates (name, type, content, parameters, default_terms, is_active) VALUES
('Standard Service Agreement', 'Service', 
'SERVICE AGREEMENT

This Service Agreement ("Agreement") is entered into on {{effective_date}} between {{company_name}} ("Company") and {{client_name}} ("Client").

1. SERVICES
Company agrees to provide {{service_description}} as detailed in Exhibit A.

2. COMPENSATION
Client agrees to pay {{total_amount}} {{currency}} according to the payment schedule outlined in Section 3.

3. PAYMENT TERMS
{{payment_terms}}

4. TERM AND TERMINATION
This Agreement shall commence on {{start_date}} and continue until {{end_date}}, unless terminated earlier in accordance with this Agreement.

5. CONFIDENTIALITY
Both parties acknowledge that they may have access to confidential information and agree to maintain such information in strict confidence.

6. GOVERNING LAW
This Agreement shall be governed by the laws of {{governing_jurisdiction}}.

IN WITNESS WHEREOF, the parties have executed this Agreement as of the date first written above.',
'[
  {"name": "effective_date", "type": "date", "required": true, "description": "Date the agreement becomes effective"},
  {"name": "company_name", "type": "string", "required": true, "description": "Name of the service provider"},
  {"name": "client_name", "type": "string", "required": true, "description": "Name of the client"},
  {"name": "service_description", "type": "text", "required": true, "description": "Detailed description of services"},
  {"name": "total_amount", "type": "number", "required": true, "description": "Total contract value"},
  {"name": "currency", "type": "string", "required": true, "default": "USD", "description": "Currency code"},
  {"name": "payment_terms", "type": "text", "required": true, "description": "Payment terms and schedule"},
  {"name": "start_date", "type": "date", "required": true, "description": "Service start date"},
  {"name": "end_date", "type": "date", "required": true, "description": "Service end date"},
  {"name": "governing_jurisdiction", "type": "string", "required": true, "default": "Delaware", "description": "Governing law jurisdiction"}
]',
'[
  {"type": "Payment", "name": "Payment Terms", "description": "Standard payment terms", "is_mandatory": true},
  {"type": "Termination", "name": "Termination Clause", "description": "Conditions for agreement termination", "is_mandatory": true},
  {"type": "Confidentiality", "name": "Confidentiality", "description": "Protection of confidential information", "is_mandatory": true},
  {"type": "Governing", "name": "Governing Law", "description": "Applicable law and jurisdiction", "is_mandatory": true}
]',
true),

('Non-Disclosure Agreement', 'NDA',
'MUTUAL NON-DISCLOSURE AGREEMENT

This Non-Disclosure Agreement ("Agreement") is entered into on {{effective_date}} between {{party1_name}} and {{party2_name}} (each a "Party" and collectively the "Parties").

1. CONFIDENTIAL INFORMATION
For purposes of this Agreement, "Confidential Information" means {{confidential_definition}}.

2. OBLIGATIONS
Each Party agrees to:
a) Hold and maintain all Confidential Information in strict confidence
b) Not disclose Confidential Information to third parties without prior written consent
c) Use Confidential Information solely for {{permitted_purpose}}

3. TERM
This Agreement shall remain in effect for {{term_duration}} from the effective date.

4. RETURN OF MATERIALS
Upon termination, each Party shall return or destroy all Confidential Information.

5. GOVERNING LAW
This Agreement shall be governed by the laws of {{governing_jurisdiction}}.

IN WITNESS WHEREOF, the parties have executed this Agreement as of the date first written above.',
'[
  {"name": "effective_date", "type": "date", "required": true, "description": "Agreement effective date"},
  {"name": "party1_name", "type": "string", "required": true, "description": "First party name"},
  {"name": "party2_name", "type": "string", "required": true, "description": "Second party name"},
  {"name": "confidential_definition", "type": "text", "required": true, "description": "Definition of confidential information"},
  {"name": "permitted_purpose", "type": "text", "required": true, "description": "Permitted use of confidential information"},
  {"name": "term_duration", "type": "string", "required": true, "default": "2 years", "description": "Duration of confidentiality obligations"},
  {"name": "governing_jurisdiction", "type": "string", "required": true, "default": "Delaware", "description": "Governing law jurisdiction"}
]',
'[
  {"type": "Confidentiality", "name": "Mutual Confidentiality", "description": "Mutual protection of confidential information", "is_mandatory": true},
  {"type": "Termination", "name": "Return of Materials", "description": "Obligations upon termination", "is_mandatory": true},
  {"type": "Governing", "name": "Governing Law", "description": "Applicable law and jurisdiction", "is_mandatory": true}
]',
true);

-- Create initial insurance policy records for the system
INSERT INTO insurance_policies (type, policy_number, provider, coverage_amount, coverage_currency, deductible_amount, deductible_currency, premium_amount, premium_currency, effective_date, expiration_date, renewal_date, coverage_details, status) VALUES
('Professional', 'PROF-2024-001', 'TechInsure LLC', 1000000.00, 'USD', 5000.00, 'USD', 15000.00, 'USD', '2024-01-01', '2024-12-31', '2024-11-01', 
ARRAY['Professional liability', 'Errors and omissions', 'Technology services coverage'], 'Active'),
('Cyber', 'CYBER-2024-001', 'CyberSecure Inc', 500000.00, 'USD', 2500.00, 'USD', 8000.00, 'USD', '2024-01-01', '2024-12-31', '2024-11-01',
ARRAY['Data breach response', 'Business interruption', 'Cyber extortion', 'Regulatory fines'], 'Active'),
('General', 'GEN-2024-001', 'Business Shield Corp', 2000000.00, 'USD', 10000.00, 'USD', 12000.00, 'USD', '2024-01-01', '2024-12-31', '2024-11-01',
ARRAY['General liability', 'Product liability', 'Personal injury', 'Advertising injury'], 'Active');