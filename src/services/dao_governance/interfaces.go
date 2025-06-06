package dao_governance

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// GovernanceService defines the main interface for DAO governance operations
type GovernanceService interface {
	// Proposal Management
	CreateProposal(ctx context.Context, request ProposalCreationRequest) (*entities.GovernanceProposal, error)
	SubmitProposal(ctx context.Context, proposalID uuid.UUID) error
	GetProposal(ctx context.Context, proposalID uuid.UUID) (*entities.GovernanceProposal, error)
	ListProposals(ctx context.Context, filter repositories.ProposalFilter) ([]*entities.GovernanceProposal, error)
	UpdateProposal(ctx context.Context, proposalID uuid.UUID, updates ProposalUpdates) error
	CancelProposal(ctx context.Context, proposalID uuid.UUID, reason string) error
	ExecuteProposal(ctx context.Context, proposalID uuid.UUID) (*ProposalExecutionResult, error)

	// Voting Management
	CastVote(ctx context.Context, request VoteRequest) (*entities.GovernanceVote, error)
	GetVote(ctx context.Context, proposalID, voterID uuid.UUID) (*entities.GovernanceVote, error)
	GetProposalVotes(ctx context.Context, proposalID uuid.UUID) ([]*entities.GovernanceVote, error)
	GetVoteResults(ctx context.Context, proposalID uuid.UUID) (*entities.ProposalVoteResult, error)
	ChangeVote(ctx context.Context, voteID uuid.UUID, newChoice entities.VoteChoice, rationale string) error

	// Member Management
	RegisterMember(ctx context.Context, request MemberRegistrationRequest) (*entities.DAOMember, error)
	GetMember(ctx context.Context, memberID uuid.UUID) (*entities.DAOMember, error)
	GetMemberByAddress(ctx context.Context, address string) (*entities.DAOMember, error)
	UpdateMember(ctx context.Context, memberID uuid.UUID, updates MemberUpdates) error
	ListMembers(ctx context.Context, filter repositories.MemberFilter) ([]*entities.DAOMember, error)
	UpdateMemberVotingPower(ctx context.Context, memberID uuid.UUID) error

	// Delegation Management
	DelegateVotes(ctx context.Context, request DelegationRequest) (*entities.VoteDelegation, error)
	RevokeDelegation(ctx context.Context, delegationID uuid.UUID) error
	GetDelegations(ctx context.Context, memberID uuid.UUID) ([]*entities.VoteDelegation, error)
	GetDelegatedVotingPower(ctx context.Context, memberID uuid.UUID) (*entities.Money, error)

	// Treasury Integration
	CreateAllocation(ctx context.Context, request AllocationRequest) (*entities.TreasuryAllocation, error)
	ExecuteAllocation(ctx context.Context, allocationID uuid.UUID) error
	GetAllocation(ctx context.Context, allocationID uuid.UUID) (*entities.TreasuryAllocation, error)
	ListAllocations(ctx context.Context, filter repositories.AllocationFilter) ([]*entities.TreasuryAllocation, error)
	ProcessInstallmentPayments(ctx context.Context) error

	// Analytics and Reporting
	GetGovernanceMetrics(ctx context.Context, timeRange repositories.TimeRange) (*repositories.GovernanceMetrics, error)
	GetMemberParticipation(ctx context.Context, memberID uuid.UUID, timeRange repositories.TimeRange) (*repositories.MemberParticipationStats, error)
	GetVotingPowerDistribution(ctx context.Context) (*repositories.VotingPowerDistribution, error)
	GenerateGovernanceReport(ctx context.Context, request ReportRequest) (*GovernanceReport, error)

	// Configuration Management
	UpdateGovernanceConfig(ctx context.Context, config *entities.GovernanceConfig) error
	GetGovernanceConfig(ctx context.Context) (*entities.GovernanceConfig, error)
}

// VotingService handles voting-specific operations
type VotingService interface {
	ValidateVoteEligibility(ctx context.Context, proposalID, voterID uuid.UUID) (*VoteEligibility, error)
	CalculateVotingPower(ctx context.Context, memberID uuid.UUID, proposalID uuid.UUID) (*entities.Money, error)
	ProcessVoteSubmission(ctx context.Context, vote *entities.GovernanceVote) error
	TallyVotes(ctx context.Context, proposalID uuid.UUID) (*VoteTallyResult, error)
	CheckQuorumReached(ctx context.Context, proposalID uuid.UUID) (bool, error)
	DetermineProposalOutcome(ctx context.Context, proposalID uuid.UUID) (*ProposalOutcome, error)
}

// MembershipService handles member-specific operations
type MembershipService interface {
	ValidateMemberRegistration(ctx context.Context, request MemberRegistrationRequest) error
	CalculateContributionScore(ctx context.Context, memberID uuid.UUID) (float64, error)
	UpdateMemberTokenBalance(ctx context.Context, memberID uuid.UUID, balance *entities.Money) error
	ProcessVestingSchedule(ctx context.Context, memberID uuid.UUID) error
	GetMemberHistory(ctx context.Context, memberID uuid.UUID) (*MemberHistory, error)
	PromoteMember(ctx context.Context, memberID uuid.UUID, newRole entities.MemberRole) error
}

// TreasuryGovernanceService handles treasury operations under governance
type TreasuryGovernanceService interface {
	ProposeAllocation(ctx context.Context, request AllocationProposalRequest) (*entities.GovernanceProposal, error)
	ValidateAllocationRequest(ctx context.Context, request AllocationRequest) error
	ExecuteTreasuryAction(ctx context.Context, action TreasuryAction) error
	GetTreasuryBalance(ctx context.Context) (*TreasuryBalance, error)
	GetAllocationSummary(ctx context.Context, timeRange repositories.TimeRange) (*repositories.TreasuryAllocationSummary, error)
	ProcessBudgetExecution(ctx context.Context, budgetID uuid.UUID) error
}

// ProposalOrchestratorService handles end-to-end proposal workflows
type ProposalOrchestratorService interface {
	StartProposalWorkflow(ctx context.Context, request ProposalWorkflowRequest) (*ProposalWorkflow, error)
	ProcessProposalStage(ctx context.Context, proposalID uuid.UUID, stage string) error
	HandleProposalCompletion(ctx context.Context, proposalID uuid.UUID) error
	GetProposalWorkflowStatus(ctx context.Context, proposalID uuid.UUID) (*ProposalWorkflowStatus, error)
	ScheduleProposalExecution(ctx context.Context, proposalID uuid.UUID, executionTime time.Time) error
}

// BlockchainIntegrationService handles on-chain interactions
type BlockchainIntegrationService interface {
	SubmitProposalOnChain(ctx context.Context, proposal *entities.GovernanceProposal) (string, error)
	SubmitVoteOnChain(ctx context.Context, vote *entities.GovernanceVote) (string, error)
	ExecuteProposalOnChain(ctx context.Context, proposalID uuid.UUID) ([]string, error)
	GetOnChainProposalStatus(ctx context.Context, onChainID string) (*OnChainProposalStatus, error)
	SyncOnChainData(ctx context.Context) error
	GetTokenBalance(ctx context.Context, address string) (*entities.Money, error)
	GetVotingPowerAt(ctx context.Context, address string, blockNumber uint64) (*entities.Money, error)
}

// Request and Response Types

// ProposalCreationRequest represents a request to create a new proposal
type ProposalCreationRequest struct {
	Title            string                     `json:"title"`
	Description      string                     `json:"description"`
	Type             entities.ProposalType      `json:"type"`
	ProposerID       uuid.UUID                  `json:"proposer_id"`
	Actions          []entities.ProposalAction  `json:"actions"`
	Parameters       map[string]interface{}     `json:"parameters,omitempty"`
	VotingPeriod     *time.Duration             `json:"voting_period,omitempty"`
	QuorumRequired   *float64                   `json:"quorum_required,omitempty"`
	PassingThreshold *float64                   `json:"passing_threshold,omitempty"`
	ExecutionDelay   *time.Duration             `json:"execution_delay,omitempty"`
	Category         string                     `json:"category,omitempty"`
	DiscussionURL    string                     `json:"discussion_url,omitempty"`
	IPFSHash         string                     `json:"ipfs_hash,omitempty"`
	IsEmergency      bool                       `json:"is_emergency,omitempty"`
}

// ProposalUpdates represents updates to a proposal
type ProposalUpdates struct {
	Title         *string                    `json:"title,omitempty"`
	Description   *string                    `json:"description,omitempty"`
	Actions       *[]entities.ProposalAction `json:"actions,omitempty"`
	Parameters    map[string]interface{}     `json:"parameters,omitempty"`
	IPFSHash      *string                    `json:"ipfs_hash,omitempty"`
	DiscussionURL *string                    `json:"discussion_url,omitempty"`
}

// VoteRequest represents a voting request
type VoteRequest struct {
	ProposalID   uuid.UUID           `json:"proposal_id"`
	VoterID      uuid.UUID           `json:"voter_id"`
	Choice       entities.VoteChoice `json:"choice"`
	Rationale    string              `json:"rationale,omitempty"`
	VoterAddress string              `json:"voter_address,omitempty"`
	Signature    string              `json:"signature,omitempty"`
}

// MemberRegistrationRequest represents a member registration request
type MemberRegistrationRequest struct {
	Address           string                         `json:"address"`
	ENSName           string                         `json:"ens_name,omitempty"`
	Handle            string                         `json:"handle,omitempty"`
	Role              entities.MemberRole            `json:"role"`
	TokenBalance      *entities.Money                `json:"token_balance"`
	VestingSchedule   *entities.VestingSchedule      `json:"vesting_schedule,omitempty"`
	ContributionProof map[string]interface{}         `json:"contribution_proof,omitempty"`
}

// MemberUpdates represents updates to member information
type MemberUpdates struct {
	Handle            *string                        `json:"handle,omitempty"`
	Role              *entities.MemberRole           `json:"role,omitempty"`
	Status            *entities.MemberStatus         `json:"status,omitempty"`
	ContributionScore *float64                       `json:"contribution_score,omitempty"`
	VestingSchedule   *entities.VestingSchedule      `json:"vesting_schedule,omitempty"`
	Metadata          map[string]interface{}         `json:"metadata,omitempty"`
}

// DelegationRequest represents a vote delegation request
type DelegationRequest struct {
	DelegatorID  uuid.UUID                 `json:"delegator_id"`
	DelegateID   uuid.UUID                 `json:"delegate_id"`
	ProposalType *entities.ProposalType    `json:"proposal_type,omitempty"`
	ExpiresAt    *time.Time                `json:"expires_at,omitempty"`
	Signature    string                    `json:"signature,omitempty"`
}

// AllocationRequest represents a treasury allocation request
type AllocationRequest struct {
	ProposalID       uuid.UUID                      `json:"proposal_id"`
	Title            string                         `json:"title"`
	Description      string                         `json:"description"`
	Amount           *entities.Money                `json:"amount"`
	Currency         string                         `json:"currency"`
	RecipientID      uuid.UUID                      `json:"recipient_id"`
	RecipientAddress string                         `json:"recipient_address"`
	Category         string                         `json:"category"`
	InstallmentPlan  *entities.InstallmentPlan      `json:"installment_plan,omitempty"`
	Conditions       []entities.AllocationCondition `json:"conditions,omitempty"`
	Milestones       []entities.AllocationMilestone `json:"milestones,omitempty"`
}

// AllocationProposalRequest represents a request to create an allocation proposal
type AllocationProposalRequest struct {
	Title            string                         `json:"title"`
	Description      string                         `json:"description"`
	ProposerID       uuid.UUID                      `json:"proposer_id"`
	AllocationRequest AllocationRequest             `json:"allocation_request"`
	Justification    string                         `json:"justification"`
	Category         string                         `json:"category"`
}

// ReportRequest represents a governance report request
type ReportRequest struct {
	Type        string                     `json:"type"` // "activity", "financial", "participation", "performance"
	TimeRange   repositories.TimeRange     `json:"time_range"`
	Format      string                     `json:"format"` // "json", "pdf", "csv"
	Recipients  []string                   `json:"recipients,omitempty"`
	Parameters  map[string]interface{}     `json:"parameters,omitempty"`
}

// Response Types

// ProposalExecutionResult represents the result of proposal execution
type ProposalExecutionResult struct {
	ProposalID      uuid.UUID                  `json:"proposal_id"`
	Success         bool                       `json:"success"`
	ActionsExecuted []entities.ProposalAction  `json:"actions_executed"`
	Results         map[string]interface{}     `json:"results"`
	TxHashes        []string                   `json:"tx_hashes,omitempty"`
	ErrorMessage    string                     `json:"error_message,omitempty"`
	ExecutedAt      time.Time                  `json:"executed_at"`
	Gas Used        uint64                     `json:"gas_used,omitempty"`
}

// VoteEligibility represents vote eligibility information
type VoteEligibility struct {
	IsEligible    bool            `json:"is_eligible"`
	VotingPower   *entities.Money `json:"voting_power"`
	Reason        string          `json:"reason,omitempty"`
	BlockSnapshot uint64          `json:"block_snapshot"`
	AlreadyVoted  bool            `json:"already_voted"`
}

// VoteTallyResult represents the result of vote tallying
type VoteTallyResult struct {
	ProposalID        uuid.UUID       `json:"proposal_id"`
	TotalVotingPower  *entities.Money `json:"total_voting_power"`
	VotesFor          *entities.Money `json:"votes_for"`
	VotesAgainst      *entities.Money `json:"votes_against"`
	VotesAbstain      *entities.Money `json:"votes_abstain"`
	VoterCount        int             `json:"voter_count"`
	QuorumReached     bool            `json:"quorum_reached"`
	QuorumRequired    *entities.Money `json:"quorum_required"`
	PassingThreshold  float64         `json:"passing_threshold"`
	Passed            bool            `json:"passed"`
	TalliedAt         time.Time       `json:"tallied_at"`
}

// ProposalOutcome represents the final outcome of a proposal
type ProposalOutcome struct {
	ProposalID       uuid.UUID                 `json:"proposal_id"`
	FinalStatus      entities.ProposalStatus   `json:"final_status"`
	VoteResults      *VoteTallyResult          `json:"vote_results"`
	ExecutionResults *ProposalExecutionResult  `json:"execution_results,omitempty"`
	Reason           string                    `json:"reason"`
	CompletedAt      time.Time                 `json:"completed_at"`
}

// MemberHistory represents a member's historical activity
type MemberHistory struct {
	MemberID            uuid.UUID                          `json:"member_id"`
	ProposalsSubmitted  []*entities.GovernanceProposal     `json:"proposals_submitted"`
	VotesCast           []*entities.GovernanceVote         `json:"votes_cast"`
	DelegationsReceived []*entities.VoteDelegation         `json:"delegations_received"`
	DelegationsMade     []*entities.VoteDelegation         `json:"delegations_made"`
	ContributionHistory []ContributionRecord               `json:"contribution_history"`
	RoleChanges         []RoleChangeRecord                 `json:"role_changes"`
	TokenTransactions   []TokenTransaction                 `json:"token_transactions"`
}

// ContributionRecord represents a contribution record
type ContributionRecord struct {
	Date         time.Time              `json:"date"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Impact       float64                `json:"impact"`
	Evidence     map[string]interface{} `json:"evidence,omitempty"`
}

// RoleChangeRecord represents a role change record
type RoleChangeRecord struct {
	Date        time.Time           `json:"date"`
	PreviousRole entities.MemberRole `json:"previous_role"`
	NewRole     entities.MemberRole `json:"new_role"`
	Reason      string              `json:"reason"`
	ChangedBy   uuid.UUID           `json:"changed_by"`
}

// TokenTransaction represents a token transaction record
type TokenTransaction struct {
	Date        time.Time       `json:"date"`
	Type        string          `json:"type"` // "mint", "burn", "transfer", "vest", "delegate"
	Amount      *entities.Money `json:"amount"`
	From        string          `json:"from,omitempty"`
	To          string          `json:"to,omitempty"`
	TxHash      string          `json:"tx_hash,omitempty"`
	Description string          `json:"description"`
}

// TreasuryAction represents an action to be performed on the treasury
type TreasuryAction struct {
	Type        string                 `json:"type"` // "transfer", "allocation", "budget", "investment"
	Target      string                 `json:"target"`
	Amount      *entities.Money        `json:"amount"`
	Currency    string                 `json:"currency"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Description string                 `json:"description"`
}

// TreasuryBalance represents current treasury balances
type TreasuryBalance struct {
	TotalValue    *entities.Money            `json:"total_value"`
	Assets        map[string]*entities.Money `json:"assets"`
	PendingOut    *entities.Money            `json:"pending_out"`
	Reserved      *entities.Money            `json:"reserved"`
	Available     *entities.Money            `json:"available"`
	LastUpdated   time.Time                  `json:"last_updated"`
}

// ProposalWorkflowRequest represents a request to start a proposal workflow
type ProposalWorkflowRequest struct {
	ProposalID     uuid.UUID              `json:"proposal_id"`
	WorkflowType   string                 `json:"workflow_type"`
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
	AutoAdvance    bool                   `json:"auto_advance"`
	Notifications  []string               `json:"notifications,omitempty"`
}

// ProposalWorkflow represents a proposal workflow
type ProposalWorkflow struct {
	ID            uuid.UUID                 `json:"id"`
	ProposalID    uuid.UUID                 `json:"proposal_id"`
	Type          string                    `json:"type"`
	CurrentStage  string                    `json:"current_stage"`
	Stages        []WorkflowStage           `json:"stages"`
	Configuration map[string]interface{}    `json:"configuration"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

// WorkflowStage represents a stage in a proposal workflow
type WorkflowStage struct {
	Name           string                 `json:"name"`
	Status         string                 `json:"status"` // "pending", "active", "completed", "failed"
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	Duration       *time.Duration         `json:"duration,omitempty"`
	Requirements   []string               `json:"requirements,omitempty"`
	Results        map[string]interface{} `json:"results,omitempty"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
}

// ProposalWorkflowStatus represents the status of a proposal workflow
type ProposalWorkflowStatus struct {
	WorkflowID      uuid.UUID       `json:"workflow_id"`
	ProposalID      uuid.UUID       `json:"proposal_id"`
	CurrentStage    string          `json:"current_stage"`
	OverallStatus   string          `json:"overall_status"`
	Progress        float64         `json:"progress"` // 0.0 to 1.0
	EstimatedCompletion *time.Time  `json:"estimated_completion,omitempty"`
	NextAction      string          `json:"next_action,omitempty"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// OnChainProposalStatus represents on-chain proposal status
type OnChainProposalStatus struct {
	OnChainID       string                  `json:"on_chain_id"`
	Status          entities.ProposalStatus `json:"status"`
	VotesFor        *entities.Money         `json:"votes_for"`
	VotesAgainst    *entities.Money         `json:"votes_against"`
	VotesAbstain    *entities.Money         `json:"votes_abstain"`
	QuorumReached   bool                    `json:"quorum_reached"`
	Passed          bool                    `json:"passed"`
	ExecutionTime   *time.Time              `json:"execution_time,omitempty"`
	BlockNumber     uint64                  `json:"block_number"`
	TransactionHash string                  `json:"transaction_hash"`
	LastSynced      time.Time               `json:"last_synced"`
}

// GovernanceReport represents a comprehensive governance report
type GovernanceReport struct {
	ID          uuid.UUID                  `json:"id"`
	Type        string                     `json:"type"`
	TimeRange   repositories.TimeRange     `json:"time_range"`
	GeneratedAt time.Time                  `json:"generated_at"`
	Summary     GovernanceReportSummary    `json:"summary"`
	Sections    []GovernanceReportSection  `json:"sections"`
	Attachments []ReportAttachment         `json:"attachments,omitempty"`
	Format      string                     `json:"format"`
}

// GovernanceReportSummary represents the summary section of a governance report
type GovernanceReportSummary struct {
	TotalProposals     int             `json:"total_proposals"`
	PassedProposals    int             `json:"passed_proposals"`
	RejectedProposals  int             `json:"rejected_proposals"`
	ActiveMembers      int             `json:"active_members"`
	ParticipationRate  float64         `json:"participation_rate"`
	TreasuryBalance    *entities.Money `json:"treasury_balance"`
	AllocationsTotal   *entities.Money `json:"allocations_total"`
	KeyInsights        []string        `json:"key_insights"`
}

// GovernanceReportSection represents a section of a governance report
type GovernanceReportSection struct {
	Title   string                 `json:"title"`
	Content interface{}            `json:"content"`
	Charts  []ChartData            `json:"charts,omitempty"`
	Tables  []TableData            `json:"tables,omitempty"`
}

// ChartData represents chart data for reports
type ChartData struct {
	Type        string                 `json:"type"` // "line", "bar", "pie", "area"
	Title       string                 `json:"title"`
	Data        interface{}            `json:"data"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// TableData represents table data for reports
type TableData struct {
	Title   string              `json:"title"`
	Headers []string            `json:"headers"`
	Rows    [][]interface{}     `json:"rows"`
	Config  map[string]interface{} `json:"config,omitempty"`
}

// ReportAttachment represents a report attachment
type ReportAttachment struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Size        int64     `json:"size"`
	URL         string    `json:"url"`
	Hash        string    `json:"hash,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}