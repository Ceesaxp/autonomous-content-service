-- Self-Improvement System Database Schema

-- Learning Artifacts Table
CREATE TABLE learning_artifacts (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    category VARCHAR(100) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    source VARCHAR(50) NOT NULL,
    source_id VARCHAR(255),
    evidence JSONB NOT NULL DEFAULT '[]',
    confidence FLOAT NOT NULL DEFAULT 0.0,
    impact_score FLOAT NOT NULL DEFAULT 0.0,
    tags JSONB NOT NULL DEFAULT '[]',
    related_artifacts JSONB NOT NULL DEFAULT '[]',
    prerequisites JSONB NOT NULL DEFAULT '[]',
    metadata JSONB NOT NULL DEFAULT '{}',
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMP WITH TIME ZONE,
    verification_date TIMESTAMP WITH TIME ZONE,
    usage_count INTEGER NOT NULL DEFAULT 0,
    success_rate FLOAT NOT NULL DEFAULT 0.0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for learning artifacts
CREATE INDEX idx_learning_artifacts_type ON learning_artifacts(type);
CREATE INDEX idx_learning_artifacts_category ON learning_artifacts(category);
CREATE INDEX idx_learning_artifacts_source ON learning_artifacts(source);
CREATE INDEX idx_learning_artifacts_status ON learning_artifacts(status);
CREATE INDEX idx_learning_artifacts_confidence ON learning_artifacts(confidence);
CREATE INDEX idx_learning_artifacts_impact ON learning_artifacts(impact_score);
CREATE INDEX idx_learning_artifacts_usage ON learning_artifacts(usage_count);
CREATE INDEX idx_learning_artifacts_created ON learning_artifacts(created_at);

-- Artifact Relationships Table (for knowledge graph)
CREATE TABLE artifact_relationships (
    id SERIAL PRIMARY KEY,
    source_id VARCHAR(255) NOT NULL REFERENCES learning_artifacts(id) ON DELETE CASCADE,
    target_id VARCHAR(255) NOT NULL REFERENCES learning_artifacts(id) ON DELETE CASCADE,
    relation_type VARCHAR(50) NOT NULL,
    weight FLOAT DEFAULT 1.0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(source_id, target_id, relation_type)
);

-- Indexes for relationships
CREATE INDEX idx_artifact_relationships_source ON artifact_relationships(source_id);
CREATE INDEX idx_artifact_relationships_target ON artifact_relationships(target_id);
CREATE INDEX idx_artifact_relationships_type ON artifact_relationships(relation_type);

-- Performance Metrics Table
CREATE TABLE performance_metrics (
    id VARCHAR(255) PRIMARY KEY,
    component VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    value FLOAT NOT NULL,
    unit VARCHAR(50),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    context JSONB NOT NULL DEFAULT '{}',
    aggregation VARCHAR(20) NOT NULL DEFAULT 'average',
    period VARCHAR(20) NOT NULL DEFAULT '1h',
    tags JSONB NOT NULL DEFAULT '[]'
);

-- Indexes for performance metrics
CREATE INDEX idx_performance_metrics_component ON performance_metrics(component);
CREATE INDEX idx_performance_metrics_name ON performance_metrics(metric_name);
CREATE INDEX idx_performance_metrics_timestamp ON performance_metrics(timestamp);
CREATE INDEX idx_performance_metrics_comp_metric ON performance_metrics(component, metric_name);
CREATE INDEX idx_performance_metrics_comp_time ON performance_metrics(component, timestamp);

-- Experiments Table
CREATE TABLE experiments (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    description TEXT,
    hypothesis TEXT NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'ab_test',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    variants JSONB NOT NULL DEFAULT '[]',
    metrics_tracked JSONB NOT NULL DEFAULT '[]',
    success_criteria JSONB NOT NULL DEFAULT '{}',
    sample_size INTEGER NOT NULL DEFAULT 0,
    current_sample INTEGER NOT NULL DEFAULT 0,
    results JSONB,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for experiments
CREATE INDEX idx_experiments_status ON experiments(status);
CREATE INDEX idx_experiments_type ON experiments(type);
CREATE INDEX idx_experiments_start_date ON experiments(start_date);
CREATE INDEX idx_experiments_created ON experiments(created_at);

-- Experiment Assignments Table
CREATE TABLE experiment_assignments (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(255) NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    entity_id VARCHAR(255) NOT NULL,
    variant_id VARCHAR(255) NOT NULL,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(experiment_id, entity_id)
);

-- Indexes for experiment assignments
CREATE INDEX idx_experiment_assignments_experiment ON experiment_assignments(experiment_id);
CREATE INDEX idx_experiment_assignments_entity ON experiment_assignments(entity_id);
CREATE INDEX idx_experiment_assignments_variant ON experiment_assignments(variant_id);

-- Experiment Conversions Table
CREATE TABLE experiment_conversions (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(255) NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    variant_id VARCHAR(255) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    metrics JSONB NOT NULL DEFAULT '{}',
    converted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for experiment conversions
CREATE INDEX idx_experiment_conversions_experiment ON experiment_conversions(experiment_id);
CREATE INDEX idx_experiment_conversions_variant ON experiment_conversions(variant_id);
CREATE INDEX idx_experiment_conversions_converted ON experiment_conversions(converted_at);

-- Capability Gaps Table
CREATE TABLE capability_gaps (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    requested_by JSONB NOT NULL DEFAULT '[]',
    frequency INTEGER NOT NULL DEFAULT 1,
    priority FLOAT NOT NULL DEFAULT 0.0,
    estimated_impact FLOAT NOT NULL DEFAULT 0.0,
    estimated_effort FLOAT NOT NULL DEFAULT 0.0,
    potential_sources JSONB NOT NULL DEFAULT '[]',
    status VARCHAR(20) NOT NULL DEFAULT 'identified',
    resolution JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for capability gaps
CREATE INDEX idx_capability_gaps_type ON capability_gaps(type);
CREATE INDEX idx_capability_gaps_status ON capability_gaps(status);
CREATE INDEX idx_capability_gaps_priority ON capability_gaps(priority);
CREATE INDEX idx_capability_gaps_impact ON capability_gaps(estimated_impact);
CREATE INDEX idx_capability_gaps_created ON capability_gaps(created_at);

-- Prompt Optimizations Table
CREATE TABLE prompt_optimizations (
    id VARCHAR(255) PRIMARY KEY,
    component VARCHAR(100) NOT NULL,
    original_prompt TEXT NOT NULL,
    optimized_prompt TEXT NOT NULL,
    llm_provider VARCHAR(100) NOT NULL,
    model_version VARCHAR(100) NOT NULL,
    improvements JSONB NOT NULL DEFAULT '{}',
    test_results JSONB NOT NULL DEFAULT '[]',
    status VARCHAR(20) NOT NULL DEFAULT 'testing',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    activated_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for prompt optimizations
CREATE INDEX idx_prompt_optimizations_component ON prompt_optimizations(component);
CREATE INDEX idx_prompt_optimizations_status ON prompt_optimizations(status);
CREATE INDEX idx_prompt_optimizations_provider ON prompt_optimizations(llm_provider);
CREATE INDEX idx_prompt_optimizations_created ON prompt_optimizations(created_at);
CREATE INDEX idx_prompt_optimizations_activated ON prompt_optimizations(activated_at);

-- Learning Sessions Table (for tracking learning processes)
CREATE TABLE learning_sessions (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    source_id VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress',
    artifacts_created JSONB NOT NULL DEFAULT '[]',
    metrics JSONB NOT NULL DEFAULT '{}',
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB NOT NULL DEFAULT '{}'
);

-- Indexes for learning sessions
CREATE INDEX idx_learning_sessions_type ON learning_sessions(type);
CREATE INDEX idx_learning_sessions_source ON learning_sessions(source_type, source_id);
CREATE INDEX idx_learning_sessions_status ON learning_sessions(status);
CREATE INDEX idx_learning_sessions_started ON learning_sessions(started_at);

-- Improvement Tracking Table
CREATE TABLE improvements (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    component VARCHAR(100) NOT NULL,
    impact FLOAT NOT NULL DEFAULT 0.0,
    effort FLOAT NOT NULL DEFAULT 0.0,
    cost FLOAT NOT NULL DEFAULT 0.0,
    time_to_value VARCHAR(50),
    roi FLOAT,
    priority FLOAT,
    status VARCHAR(20) NOT NULL DEFAULT 'identified',
    implementation_plan JSONB,
    results JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    implemented_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for improvements
CREATE INDEX idx_improvements_type ON improvements(type);
CREATE INDEX idx_improvements_component ON improvements(component);
CREATE INDEX idx_improvements_status ON improvements(status);
CREATE INDEX idx_improvements_priority ON improvements(priority);
CREATE INDEX idx_improvements_roi ON improvements(roi);
CREATE INDEX idx_improvements_created ON improvements(created_at);

-- Competitive Intelligence Table
CREATE TABLE competitive_intelligence (
    id VARCHAR(255) PRIMARY KEY,
    competitor_name VARCHAR(255) NOT NULL,
    data_type VARCHAR(50) NOT NULL, -- pricing, features, performance, etc.
    data JSONB NOT NULL DEFAULT '{}',
    source VARCHAR(100),
    confidence FLOAT NOT NULL DEFAULT 0.0,
    collected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB NOT NULL DEFAULT '{}'
);

-- Indexes for competitive intelligence
CREATE INDEX idx_competitive_intelligence_competitor ON competitive_intelligence(competitor_name);
CREATE INDEX idx_competitive_intelligence_type ON competitive_intelligence(data_type);
CREATE INDEX idx_competitive_intelligence_collected ON competitive_intelligence(collected_at);
CREATE INDEX idx_competitive_intelligence_expires ON competitive_intelligence(expires_at);

-- System Health Monitoring Table
CREATE TABLE system_health_checks (
    id SERIAL PRIMARY KEY,
    component VARCHAR(100) NOT NULL,
    check_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    details JSONB NOT NULL DEFAULT '{}',
    response_time_ms INTEGER,
    checked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for system health
CREATE INDEX idx_system_health_component ON system_health_checks(component);
CREATE INDEX idx_system_health_type ON system_health_checks(check_type);
CREATE INDEX idx_system_health_status ON system_health_checks(status);
CREATE INDEX idx_system_health_checked ON system_health_checks(checked_at);

-- Knowledge Graph Clusters Table
CREATE TABLE knowledge_clusters (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    description TEXT,
    node_ids JSONB NOT NULL DEFAULT '[]',
    coherence FLOAT NOT NULL DEFAULT 0.0,
    cluster_type VARCHAR(50),
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for knowledge clusters
CREATE INDEX idx_knowledge_clusters_coherence ON knowledge_clusters(coherence);
CREATE INDEX idx_knowledge_clusters_type ON knowledge_clusters(cluster_type);
CREATE INDEX idx_knowledge_clusters_created ON knowledge_clusters(created_at);

-- Add some helpful views
CREATE VIEW active_learning_artifacts AS
SELECT * FROM learning_artifacts 
WHERE status = 'active' 
AND (valid_until IS NULL OR valid_until > NOW());

CREATE VIEW high_impact_artifacts AS
SELECT * FROM learning_artifacts 
WHERE status = 'active' 
AND impact_score > 0.7 
AND confidence > 0.8;

CREATE VIEW recent_metrics AS
SELECT DISTINCT ON (component, metric_name) 
    component, metric_name, value, timestamp
FROM performance_metrics 
ORDER BY component, metric_name, timestamp DESC;

CREATE VIEW experiment_summary AS
SELECT 
    id,
    name,
    status,
    type,
    sample_size,
    current_sample,
    CASE 
        WHEN current_sample > 0 THEN (current_sample::float / sample_size * 100)
        ELSE 0 
    END as completion_percentage,
    start_date,
    end_date,
    created_at
FROM experiments;

-- Performance optimization function
CREATE OR REPLACE FUNCTION update_learning_artifact_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for automatic timestamp updates
CREATE TRIGGER learning_artifacts_updated_at_trigger
    BEFORE UPDATE ON learning_artifacts
    FOR EACH ROW
    EXECUTE FUNCTION update_learning_artifact_updated_at();

CREATE TRIGGER experiments_updated_at_trigger
    BEFORE UPDATE ON experiments
    FOR EACH ROW
    EXECUTE FUNCTION update_learning_artifact_updated_at();

CREATE TRIGGER capability_gaps_updated_at_trigger
    BEFORE UPDATE ON capability_gaps
    FOR EACH ROW
    EXECUTE FUNCTION update_learning_artifact_updated_at();

CREATE TRIGGER knowledge_clusters_updated_at_trigger
    BEFORE UPDATE ON knowledge_clusters
    FOR EACH ROW
    EXECUTE FUNCTION update_learning_artifact_updated_at();

-- Add comments for documentation
COMMENT ON TABLE learning_artifacts IS 'Stores learned knowledge artifacts with metadata and relationships';
COMMENT ON TABLE artifact_relationships IS 'Stores relationships between learning artifacts for knowledge graph';
COMMENT ON TABLE performance_metrics IS 'Stores system performance metrics for analysis and optimization';
COMMENT ON TABLE experiments IS 'Stores A/B tests and experiments for system optimization';
COMMENT ON TABLE capability_gaps IS 'Tracks identified capability gaps and their resolution status';
COMMENT ON TABLE prompt_optimizations IS 'Stores prompt optimization results and configurations';
COMMENT ON TABLE competitive_intelligence IS 'Stores competitive analysis data and market intelligence';
COMMENT ON TABLE system_health_checks IS 'Monitors system health across all components';
COMMENT ON TABLE knowledge_clusters IS 'Groups related learning artifacts into coherent clusters';