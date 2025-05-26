package events

// DecisionInitiated is emitted when a new decision process begins
type DecisionInitiated struct {
	BaseEvent
	DecisionID   string                 `json:"decision_id"`
	DecisionType string                 `json:"decision_type"`
	Priority     string                 `json:"priority"`
	Context      map[string]interface{} `json:"context"`
	Requester    string                 `json:"requester"`
}

// DecisionAnalyzed is emitted when decision analysis is complete
type DecisionAnalyzed struct {
	BaseEvent
	DecisionID       string   `json:"decision_id"`
	OptionsEvaluated int      `json:"options_evaluated"`
	TopOptions       []string `json:"top_options"`
	ConfidenceScore  float64  `json:"confidence_score"`
	AnalysisDuration int64    `json:"analysis_duration_ms"`
}

// DecisionMade is emitted when a decision is finalized
type DecisionMade struct {
	BaseEvent
	DecisionID       string  `json:"decision_id"`
	SelectedOptionID string  `json:"selected_option_id"`
	Justification    string  `json:"justification"`
	ConfidenceScore  float64 `json:"confidence_score"`
	AutoApproved     bool    `json:"auto_approved"`
}

// DecisionOverridden is emitted when a decision is manually overridden
type DecisionOverridden struct {
	BaseEvent
	DecisionID       string `json:"decision_id"`
	OriginalOptionID string `json:"original_option_id"`
	OverrideOptionID string `json:"override_option_id"`
	OverrideReason   string `json:"override_reason"`
	AuthorizedBy     string `json:"authorized_by"`
}

// DecisionExecuted is emitted when a decision is put into action
type DecisionExecuted struct {
	BaseEvent
	DecisionID    string                 `json:"decision_id"`
	Success       bool                   `json:"success"`
	ExecutionTime int64                  `json:"execution_time_ms"`
	Results       map[string]interface{} `json:"results"`
	SideEffects   []string               `json:"side_effects"`
}

// DecisionReverted is emitted when a decision is rolled back
type DecisionReverted struct {
	BaseEvent
	DecisionID    string `json:"decision_id"`
	RevertReason  string `json:"revert_reason"`
	RevertSuccess bool   `json:"revert_success"`
	Impact        string `json:"impact"`
}

// PolicyViolationDetected is emitted when a decision violates policies
type PolicyViolationDetected struct {
	BaseEvent
	DecisionID  string   `json:"decision_id"`
	PolicyID    string   `json:"policy_id"`
	PolicyName  string   `json:"policy_name"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Violations  []string `json:"violations"`
}

// EthicalConcernRaised is emitted when ethical issues are detected
type EthicalConcernRaised struct {
	BaseEvent
	DecisionID   string   `json:"decision_id"`
	GuidelineID  string   `json:"guideline_id"`
	ConcernLevel string   `json:"concern_level"`
	Description  string   `json:"description"`
	RedLines     []string `json:"red_lines"`
}

// EmergencyProtocolActivated is emitted during critical failures
type EmergencyProtocolActivated struct {
	BaseEvent
	TriggerType   string                 `json:"trigger_type"`
	Severity      string                 `json:"severity"`
	AffectedAreas []string               `json:"affected_areas"`
	Actions       []string               `json:"actions"`
	SystemState   map[string]interface{} `json:"system_state"`
}

// DecisionQualityAssessed is emitted after decision outcome evaluation
type DecisionQualityAssessed struct {
	BaseEvent
	DecisionID       string   `json:"decision_id"`
	QualityScore     float64  `json:"quality_score"`
	ExpectedOutcome  string   `json:"expected_outcome"`
	ActualOutcome    string   `json:"actual_outcome"`
	LessonsLearned   []string `json:"lessons_learned"`
	ImprovementAreas []string `json:"improvement_areas"`
}

// StakeholderImpactAssessed is emitted after stakeholder analysis
type StakeholderImpactAssessed struct {
	BaseEvent
	DecisionID         string                 `json:"decision_id"`
	StakeholderGroups  []string               `json:"stakeholder_groups"`
	OverallSentiment   float64                `json:"overall_sentiment"`
	ImpactSummary      map[string]interface{} `json:"impact_summary"`
	MitigationRequired bool                   `json:"mitigation_required"`
}

// DecisionConflictDetected is emitted when competing priorities clash
type DecisionConflictDetected struct {
	BaseEvent
	ConflictID       string   `json:"conflict_id"`
	DecisionIDs      []string `json:"decision_ids"`
	ConflictType     string   `json:"conflict_type"`
	ConflictSeverity string   `json:"conflict_severity"`
	ResolutionNeeded bool     `json:"resolution_needed"`
}

// DecisionThresholdReached is emitted when override thresholds are hit
type DecisionThresholdReached struct {
	BaseEvent
	DecisionID     string  `json:"decision_id"`
	ThresholdType  string  `json:"threshold_type"`
	ThresholdValue float64 `json:"threshold_value"`
	ActualValue    float64 `json:"actual_value"`
	ActionRequired string  `json:"action_required"`
}

// DecisionAuditCreated is emitted when audit trail is generated
type DecisionAuditCreated struct {
	BaseEvent
	DecisionID      string  `json:"decision_id"`
	AuditType       string  `json:"audit_type"`
	AuditPeriod     string  `json:"audit_period"`
	FindingsCount   int     `json:"findings_count"`
	ComplianceScore float64 `json:"compliance_score"`
}
