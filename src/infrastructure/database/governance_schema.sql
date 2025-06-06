-- DAO Governance Database Schema

-- ENUM types for governance
DO $$ BEGIN
    CREATE TYPE proposal_type AS ENUM ('Treasury', 'Parameter', 'Upgrade', 'Emergency', 'Membership', 'Policy');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE proposal_status AS ENUM ('Draft', 'Submitted', 'Active', 'Passed', 'Rejected', 'Executed', 'Canceled', 'Expired');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE vote_choice AS ENUM ('For', 'Against', 'Abstain');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE member_role AS ENUM ('Founder', 'Core', 'Contributor', 'Delegee', 'Observer');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE member_status AS ENUM ('Active', 'Inactive', 'Suspended', 'Removed');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE allocation_status AS ENUM ('Pending', 'Approved', 'Disbursed', 'Completed', 'Canceled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Governance Proposals
CREATE TABLE IF NOT EXISTS governance_proposals (
    proposal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    type proposal_type NOT NULL DEFAULT 'Treasury',
    status proposal_status NOT NULL DEFAULT 'Draft',
    proposer_id UUID NOT NULL,
    proposer_address VARCHAR(42) NOT NULL,
    voting_power_amount BIGINT DEFAULT 0,
    voting_power_currency VARCHAR(10) DEFAULT 'TOKEN',
    quorum_required DECIMAL(5,4) NOT NULL DEFAULT 0.04, -- 4%
    passing_threshold DECIMAL(5,4) NOT NULL DEFAULT 0.51, -- 51%
    voting_start_time TIMESTAMP WITH TIME ZONE,
    voting_end_time TIMESTAMP WITH TIME ZONE,
    execution_delay INTERVAL DEFAULT '24 hours',
    execution_deadline TIMESTAMP WITH TIME ZONE,
    parameters JSONB DEFAULT '{}',
    actions JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    ipfs_hash VARCHAR(64),
    on_chain_proposal_id VARCHAR(128),
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    executed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Governance Votes
CREATE TABLE IF NOT EXISTS governance_votes (
    vote_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    proposal_id UUID NOT NULL REFERENCES governance_proposals(proposal_id) ON DELETE CASCADE,
    voter_id UUID NOT NULL,
    voter_address VARCHAR(42) NOT NULL,
    choice vote_choice NOT NULL,
    voting_power_amount BIGINT NOT NULL DEFAULT 0,
    voting_power_currency VARCHAR(10) DEFAULT 'TOKEN',
    weight DECIMAL(3,2) NOT NULL DEFAULT 1.00,
    delegated_from JSONB DEFAULT '[]',
    rationale TEXT,
    on_chain_tx_hash VARCHAR(66),
    signature TEXT,
    metadata JSONB DEFAULT '{}',
    voted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(proposal_id, voter_id)
);

-- DAO Members
CREATE TABLE IF NOT EXISTS dao_members (
    member_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    address VARCHAR(42) UNIQUE NOT NULL,
    ens_name VARCHAR(255),
    handle VARCHAR(50) UNIQUE,
    role member_role NOT NULL DEFAULT 'Observer',
    status member_status NOT NULL DEFAULT 'Active',
    token_balance_amount BIGINT DEFAULT 0,
    token_balance_currency VARCHAR(10) DEFAULT 'TOKEN',
    voting_power_amount BIGINT DEFAULT 0,
    voting_power_currency VARCHAR(10) DEFAULT 'TOKEN',
    delegated_power_amount BIGINT DEFAULT 0,
    delegated_power_currency VARCHAR(10) DEFAULT 'TOKEN',
    delegated_to UUID,
    contribution_score DECIMAL(10,2) DEFAULT 0.00,
    proposals_submitted INTEGER DEFAULT 0,
    votes_participated INTEGER DEFAULT 0,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    vesting_schedule JSONB,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Treasury Allocations
CREATE TABLE IF NOT EXISTS treasury_allocations (
    allocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    proposal_id UUID NOT NULL REFERENCES governance_proposals(proposal_id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    amount_value BIGINT NOT NULL,
    amount_currency VARCHAR(10) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    recipient_id UUID NOT NULL,
    recipient_address VARCHAR(42) NOT NULL,
    category VARCHAR(50) NOT NULL, -- 'operations', 'development', 'marketing', 'rewards'
    status allocation_status NOT NULL DEFAULT 'Pending',
    installment_plan JSONB,
    conditions JSONB DEFAULT '[]',
    milestones JSONB DEFAULT '[]',
    approved_at TIMESTAMP WITH TIME ZONE,
    disbursed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Vote Delegations
CREATE TABLE IF NOT EXISTS vote_delegations (
    delegation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delegator_id UUID NOT NULL REFERENCES dao_members(member_id) ON DELETE CASCADE,
    delegate_id UUID NOT NULL REFERENCES dao_members(member_id) ON DELETE CASCADE,
    proposal_type proposal_type, -- NULL means all proposal types
    voting_power_amount BIGINT NOT NULL DEFAULT 0,
    voting_power_currency VARCHAR(10) DEFAULT 'TOKEN',
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(delegator_id, delegate_id, proposal_type)
);

-- Governance Configuration
CREATE TABLE IF NOT EXISTS governance_config (
    config_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    proposal_threshold_amount BIGINT DEFAULT 1000000000000000000, -- 1 token (18 decimals)
    proposal_threshold_currency VARCHAR(10) DEFAULT 'TOKEN',
    quorum_percentage DECIMAL(5,2) NOT NULL DEFAULT 4.00, -- 4%
    passing_threshold DECIMAL(5,4) NOT NULL DEFAULT 0.51, -- 51%
    voting_period INTERVAL NOT NULL DEFAULT '7 days',
    execution_delay INTERVAL NOT NULL DEFAULT '24 hours',
    max_actions INTEGER DEFAULT 10,
    token_address VARCHAR(42),
    timelock_address VARCHAR(42),
    treasury_address VARCHAR(42),
    allow_delegation BOOLEAN DEFAULT TRUE,
    require_reason BOOLEAN DEFAULT FALSE,
    emergency_pause_enabled BOOLEAN DEFAULT TRUE,
    parameters JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Governance Events
CREATE TABLE IF NOT EXISTS governance_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(100) NOT NULL,
    proposal_id UUID REFERENCES governance_proposals(proposal_id),
    actor_id UUID NOT NULL,
    data JSONB NOT NULL DEFAULT '{}',
    tx_hash VARCHAR(66),
    block_hash VARCHAR(66),
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance

-- Governance Proposals indexes
CREATE INDEX IF NOT EXISTS idx_governance_proposals_status ON governance_proposals(status);
CREATE INDEX IF NOT EXISTS idx_governance_proposals_type ON governance_proposals(type);
CREATE INDEX IF NOT EXISTS idx_governance_proposals_proposer ON governance_proposals(proposer_id);
CREATE INDEX IF NOT EXISTS idx_governance_proposals_voting_period ON governance_proposals(voting_start_time, voting_end_time);
CREATE INDEX IF NOT EXISTS idx_governance_proposals_created_at ON governance_proposals(created_at DESC);

-- Governance Votes indexes
CREATE INDEX IF NOT EXISTS idx_governance_votes_proposal ON governance_votes(proposal_id);
CREATE INDEX IF NOT EXISTS idx_governance_votes_voter ON governance_votes(voter_id);
CREATE INDEX IF NOT EXISTS idx_governance_votes_choice ON governance_votes(choice);
CREATE INDEX IF NOT EXISTS idx_governance_votes_voted_at ON governance_votes(voted_at DESC);

-- DAO Members indexes
CREATE INDEX IF NOT EXISTS idx_dao_members_address ON dao_members(address);
CREATE INDEX IF NOT EXISTS idx_dao_members_role ON dao_members(role);
CREATE INDEX IF NOT EXISTS idx_dao_members_status ON dao_members(status);
CREATE INDEX IF NOT EXISTS idx_dao_members_contribution_score ON dao_members(contribution_score DESC);
CREATE INDEX IF NOT EXISTS idx_dao_members_joined_at ON dao_members(joined_at DESC);
CREATE INDEX IF NOT EXISTS idx_dao_members_last_activity ON dao_members(last_activity DESC);

-- Treasury Allocations indexes
CREATE INDEX IF NOT EXISTS idx_treasury_allocations_proposal ON treasury_allocations(proposal_id);
CREATE INDEX IF NOT EXISTS idx_treasury_allocations_recipient ON treasury_allocations(recipient_id);
CREATE INDEX IF NOT EXISTS idx_treasury_allocations_status ON treasury_allocations(status);
CREATE INDEX IF NOT EXISTS idx_treasury_allocations_category ON treasury_allocations(category);
CREATE INDEX IF NOT EXISTS idx_treasury_allocations_created_at ON treasury_allocations(created_at DESC);

-- Vote Delegations indexes
CREATE INDEX IF NOT EXISTS idx_vote_delegations_delegator ON vote_delegations(delegator_id);
CREATE INDEX IF NOT EXISTS idx_vote_delegations_delegate ON vote_delegations(delegate_id);
CREATE INDEX IF NOT EXISTS idx_vote_delegations_active ON vote_delegations(is_active);
CREATE INDEX IF NOT EXISTS idx_vote_delegations_proposal_type ON vote_delegations(proposal_type);

-- Governance Events indexes
CREATE INDEX IF NOT EXISTS idx_governance_events_type ON governance_events(type);
CREATE INDEX IF NOT EXISTS idx_governance_events_proposal ON governance_events(proposal_id);
CREATE INDEX IF NOT EXISTS idx_governance_events_actor ON governance_events(actor_id);
CREATE INDEX IF NOT EXISTS idx_governance_events_occurred_at ON governance_events(occurred_at DESC);

-- Triggers for updating timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers to governance tables
DROP TRIGGER IF EXISTS update_governance_proposals_updated_at ON governance_proposals;
CREATE TRIGGER update_governance_proposals_updated_at 
    BEFORE UPDATE ON governance_proposals 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_governance_votes_updated_at ON governance_votes;
CREATE TRIGGER update_governance_votes_updated_at 
    BEFORE UPDATE ON governance_votes 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_dao_members_updated_at ON dao_members;
CREATE TRIGGER update_dao_members_updated_at 
    BEFORE UPDATE ON dao_members 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_treasury_allocations_updated_at ON treasury_allocations;
CREATE TRIGGER update_treasury_allocations_updated_at 
    BEFORE UPDATE ON treasury_allocations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_vote_delegations_updated_at ON vote_delegations;
CREATE TRIGGER update_vote_delegations_updated_at 
    BEFORE UPDATE ON vote_delegations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_governance_config_updated_at ON governance_config;
CREATE TRIGGER update_governance_config_updated_at 
    BEFORE UPDATE ON governance_config 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default governance configuration
INSERT INTO governance_config (
    config_id,
    name,
    proposal_threshold_amount,
    quorum_percentage,
    passing_threshold,
    voting_period,
    execution_delay,
    max_actions,
    allow_delegation,
    require_reason,
    emergency_pause_enabled,
    parameters
) VALUES (
    'default',
    'Default DAO Configuration',
    1000000000000000000, -- 1 TOKEN (18 decimals)
    4.00, -- 4% quorum
    0.51, -- 51% passing threshold
    '7 days'::interval,
    '24 hours'::interval,
    10,
    TRUE,
    FALSE,
    TRUE,
    '{
        "min_voting_delay": "1 hour",
        "max_voting_delay": "7 days",
        "min_voting_period": "1 day",
        "max_voting_period": "14 days",
        "emergency_voting_period": "1 day",
        "timelock_delay": "24 hours",
        "emergency_timelock_delay": "2 hours"
    }'::jsonb
) ON CONFLICT (config_id) DO UPDATE SET
    updated_at = NOW();

-- Sample DAO founder member (will be updated with real data)
INSERT INTO dao_members (
    member_id,
    address,
    handle,
    role,
    status,
    token_balance_amount,
    voting_power_amount,
    contribution_score,
    proposals_submitted,
    votes_participated,
    metadata
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    '0x0000000000000000000000000000000000000001',
    'founder',
    'Founder',
    'Active',
    100000000000000000000, -- 100 tokens
    100000000000000000000, -- 100 voting power
    100.00,
    0,
    0,
    '{"initial_member": true, "role": "system"}'::jsonb
) ON CONFLICT (address) DO UPDATE SET
    updated_at = NOW();

-- Sample proposal types configuration in metadata
-- This could be moved to a separate table if needed
UPDATE governance_config SET parameters = jsonb_set(
    parameters,
    '{proposal_types}',
    '[
        {
            "id": 1,
            "name": "Treasury",
            "quorum_fraction": 400,
            "voting_period": "5 days",
            "proposal_threshold": 1000,
            "requires_timelock": true,
            "emergency_bypass": false
        },
        {
            "id": 2,
            "name": "Parameter",
            "quorum_fraction": 300,
            "voting_period": "3 days",
            "proposal_threshold": 500,
            "requires_timelock": true,
            "emergency_bypass": false
        },
        {
            "id": 3,
            "name": "Emergency",
            "quorum_fraction": 600,
            "voting_period": "1 day",
            "proposal_threshold": 5000,
            "requires_timelock": false,
            "emergency_bypass": true
        },
        {
            "id": 4,
            "name": "Upgrade",
            "quorum_fraction": 800,
            "voting_period": "7 days",
            "proposal_threshold": 2000,
            "requires_timelock": true,
            "emergency_bypass": false
        }
    ]'::jsonb
) WHERE config_id = 'default';

-- Views for common queries

-- Active members view
CREATE OR REPLACE VIEW active_dao_members AS
SELECT 
    member_id,
    address,
    ens_name,
    handle,
    role,
    token_balance_amount,
    voting_power_amount,
    contribution_score,
    proposals_submitted,
    votes_participated,
    last_activity,
    joined_at
FROM dao_members 
WHERE status = 'Active';

-- Active proposals view
CREATE OR REPLACE VIEW active_proposals AS
SELECT 
    proposal_id,
    title,
    description,
    type,
    status,
    proposer_id,
    quorum_required,
    passing_threshold,
    voting_start_time,
    voting_end_time,
    submitted_at,
    created_at
FROM governance_proposals 
WHERE status IN ('Active', 'Submitted') 
    AND voting_end_time > NOW();

-- Proposal vote summary view
CREATE OR REPLACE VIEW proposal_vote_summary AS
SELECT 
    p.proposal_id,
    p.title,
    p.status,
    p.quorum_required,
    p.passing_threshold,
    COUNT(v.vote_id) as total_votes,
    SUM(CASE WHEN v.choice = 'For' THEN v.voting_power_amount ELSE 0 END) as votes_for,
    SUM(CASE WHEN v.choice = 'Against' THEN v.voting_power_amount ELSE 0 END) as votes_against,
    SUM(CASE WHEN v.choice = 'Abstain' THEN v.voting_power_amount ELSE 0 END) as votes_abstain,
    SUM(v.voting_power_amount) as total_voting_power
FROM governance_proposals p
LEFT JOIN governance_votes v ON p.proposal_id = v.proposal_id
GROUP BY p.proposal_id, p.title, p.status, p.quorum_required, p.passing_threshold;

-- Member activity summary view
CREATE OR REPLACE VIEW member_activity_summary AS
SELECT 
    m.member_id,
    m.address,
    m.handle,
    m.role,
    m.contribution_score,
    COUNT(DISTINCT p.proposal_id) as proposals_created,
    COUNT(DISTINCT v.vote_id) as votes_cast,
    COUNT(DISTINCT d_out.delegation_id) as delegations_made,
    COUNT(DISTINCT d_in.delegation_id) as delegations_received,
    m.last_activity,
    m.joined_at
FROM dao_members m
LEFT JOIN governance_proposals p ON m.member_id = p.proposer_id
LEFT JOIN governance_votes v ON m.member_id = v.voter_id
LEFT JOIN vote_delegations d_out ON m.member_id = d_out.delegator_id AND d_out.is_active = TRUE
LEFT JOIN vote_delegations d_in ON m.member_id = d_in.delegate_id AND d_in.is_active = TRUE
WHERE m.status = 'Active'
GROUP BY m.member_id, m.address, m.handle, m.role, m.contribution_score, m.last_activity, m.joined_at;

COMMENT ON TABLE governance_proposals IS 'DAO governance proposals with voting and execution tracking';
COMMENT ON TABLE governance_votes IS 'Individual votes cast on governance proposals';
COMMENT ON TABLE dao_members IS 'DAO members with roles, voting power, and activity tracking';
COMMENT ON TABLE treasury_allocations IS 'Treasury fund allocations approved through governance';
COMMENT ON TABLE vote_delegations IS 'Vote delegation relationships between DAO members';
COMMENT ON TABLE governance_config IS 'DAO governance configuration and parameters';
COMMENT ON TABLE governance_events IS 'Audit log of all governance-related events';