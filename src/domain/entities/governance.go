package entities

import (
	"time"

	"github.com/google/uuid"
)

// ProposalType represents the type of governance proposal
type ProposalType string

const (
	ProposalTypeTreasury   ProposalType = "Treasury"
	ProposalTypeParameter  ProposalType = "Parameter"
	ProposalTypeUpgrade    ProposalType = "Upgrade"
	ProposalTypeEmergency  ProposalType = "Emergency"
	ProposalTypeMembership ProposalType = "Membership"
	ProposalTypePolicy     ProposalType = "Policy"
)

// ProposalStatus represents the current status of a proposal
type ProposalStatus string

const (
	ProposalStatusDraft     ProposalStatus = "Draft"
	ProposalStatusSubmitted ProposalStatus = "Submitted"
	ProposalStatusActive    ProposalStatus = "Active"
	ProposalStatusPassed    ProposalStatus = "Passed"
	ProposalStatusRejected  ProposalStatus = "Rejected"
	ProposalStatusExecuted  ProposalStatus = "Executed"
	ProposalStatusCanceled  ProposalStatus = "Canceled"
	ProposalStatusExpired   ProposalStatus = "Expired"
)

// VoteChoice represents a member's vote choice
type VoteChoice string

const (
	VoteChoiceFor     VoteChoice = "For"
	VoteChoiceAgainst VoteChoice = "Against"
	VoteChoiceAbstain VoteChoice = "Abstain"
)

// MemberRole represents the role of a DAO member
type MemberRole string

const (
	MemberRoleFounder     MemberRole = "Founder"
	MemberRoleCore        MemberRole = "Core"
	MemberRoleContributor MemberRole = "Contributor"
	MemberRoleDelegee     MemberRole = "Delegee"
	MemberRoleObserver    MemberRole = "Observer"
)

// MemberStatus represents the status of a DAO member
type MemberStatus string

const (
	MemberStatusActive   MemberStatus = "Active"
	MemberStatusInactive MemberStatus = "Inactive"
	MemberStatusSuspended MemberStatus = "Suspended"
	MemberStatusRemoved  MemberStatus = "Removed"
)

// AllocationStatus represents the status of a treasury allocation
type AllocationStatus string

const (
	AllocationStatusPending    AllocationStatus = "Pending"
	AllocationStatusApproved   AllocationStatus = "Approved"
	AllocationStatusDisbursed  AllocationStatus = "Disbursed"
	AllocationStatusCompleted  AllocationStatus = "Completed"
	AllocationStatusCanceled   AllocationStatus = "Canceled"
)

// GovernanceProposal represents a DAO governance proposal
type GovernanceProposal struct {
	ID                uuid.UUID              `json:"id" db:"proposal_id"`
	Title             string                 `json:"title" db:"title"`
	Description       string                 `json:"description" db:"description"`
	Type              ProposalType           `json:"type" db:"type"`
	Status            ProposalStatus         `json:"status" db:"status"`
	ProposerID        uuid.UUID              `json:"proposer_id" db:"proposer_id"`
	ProposerAddress   string                 `json:"proposer_address" db:"proposer_address"`
	VotingPower       *Money                 `json:"voting_power,omitempty" db:"voting_power"`
	QuorumRequired    float64                `json:"quorum_required" db:"quorum_required"`
	PassingThreshold  float64                `json:"passing_threshold" db:"passing_threshold"`
	VotingStartTime   time.Time              `json:"voting_start_time" db:"voting_start_time"`
	VotingEndTime     time.Time              `json:"voting_end_time" db:"voting_end_time"`
	ExecutionDelay    time.Duration          `json:"execution_delay" db:"execution_delay"`
	ExecutionDeadline *time.Time             `json:"execution_deadline,omitempty" db:"execution_deadline"`
	Parameters        map[string]interface{} `json:"parameters,omitempty" db:"parameters"`
	Actions           []ProposalAction       `json:"actions,omitempty" db:"actions"`
	Metadata          map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	IPFSHash          string                 `json:"ipfs_hash,omitempty" db:"ipfs_hash"`
	OnChainProposalID *string                `json:"on_chain_proposal_id,omitempty" db:"on_chain_proposal_id"`
	SubmittedAt       time.Time              `json:"submitted_at" db:"submitted_at"`
	ExecutedAt        *time.Time             `json:"executed_at,omitempty" db:"executed_at"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// ProposalAction represents an action to be executed if proposal passes
type ProposalAction struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"` // "transfer", "parameter_change", "contract_call"
	Target       string                 `json:"target"`
	Method       string                 `json:"method,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Value        *Money                 `json:"value,omitempty"`
	Description  string                 `json:"description"`
	ExecutionOrder int                  `json:"execution_order"`
}

// GovernanceVote represents a vote on a proposal
type GovernanceVote struct {
	ID               uuid.UUID              `json:"id" db:"vote_id"`
	ProposalID       uuid.UUID              `json:"proposal_id" db:"proposal_id"`
	VoterID          uuid.UUID              `json:"voter_id" db:"voter_id"`
	VoterAddress     string                 `json:"voter_address" db:"voter_address"`
	Choice           VoteChoice             `json:"choice" db:"choice"`
	VotingPower      *Money                 `json:"voting_power" db:"voting_power"`
	Weight           float64                `json:"weight" db:"weight"`
	DelegatedFrom    []uuid.UUID            `json:"delegated_from,omitempty" db:"delegated_from"`
	Rationale        string                 `json:"rationale,omitempty" db:"rationale"`
	OnChainTxHash    string                 `json:"on_chain_tx_hash,omitempty" db:"on_chain_tx_hash"`
	Signature        string                 `json:"signature,omitempty" db:"signature"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	VotedAt          time.Time              `json:"voted_at" db:"voted_at"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// DAOMember represents a member of the DAO
type DAOMember struct {
	ID                 uuid.UUID              `json:"id" db:"member_id"`
	Address            string                 `json:"address" db:"address"`
	ENSName            string                 `json:"ens_name,omitempty" db:"ens_name"`
	Handle             string                 `json:"handle,omitempty" db:"handle"`
	Role               MemberRole             `json:"role" db:"role"`
	Status             MemberStatus           `json:"status" db:"status"`
	TokenBalance       *Money                 `json:"token_balance" db:"token_balance"`
	VotingPower        *Money                 `json:"voting_power" db:"voting_power"`
	DelegatedPower     *Money                 `json:"delegated_power,omitempty" db:"delegated_power"`
	DelegatedTo        *uuid.UUID             `json:"delegated_to,omitempty" db:"delegated_to"`
	ContributionScore  float64                `json:"contribution_score" db:"contribution_score"`
	ProposalsSubmitted int                    `json:"proposals_submitted" db:"proposals_submitted"`
	VotesParticipated  int                    `json:"votes_participated" db:"votes_participated"`
	LastActivity       time.Time              `json:"last_activity" db:"last_activity"`
	JoinedAt           time.Time              `json:"joined_at" db:"joined_at"`
	VestingSchedule    *VestingSchedule       `json:"vesting_schedule,omitempty" db:"vesting_schedule"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
}

// VestingSchedule represents token vesting details for a member
type VestingSchedule struct {
	TotalAmount     *Money        `json:"total_amount"`
	VestedAmount    *Money        `json:"vested_amount"`
	ClaimedAmount   *Money        `json:"claimed_amount"`
	VestingStart    time.Time     `json:"vesting_start"`
	VestingDuration time.Duration `json:"vesting_duration"`
	CliffDuration   time.Duration `json:"cliff_duration"`
	VestingType     string        `json:"vesting_type"` // "linear", "milestone", "performance"
}

// TreasuryAllocation represents a budget allocation from treasury
type TreasuryAllocation struct {
	ID               uuid.UUID              `json:"id" db:"allocation_id"`
	ProposalID       uuid.UUID              `json:"proposal_id" db:"proposal_id"`
	Title            string                 `json:"title" db:"title"`
	Description      string                 `json:"description" db:"description"`
	Amount           *Money                 `json:"amount" db:"amount"`
	Currency         string                 `json:"currency" db:"currency"`
	RecipientID      uuid.UUID              `json:"recipient_id" db:"recipient_id"`
	RecipientAddress string                 `json:"recipient_address" db:"recipient_address"`
	Category         string                 `json:"category" db:"category"` // "operations", "development", "marketing", "rewards"
	Status           AllocationStatus       `json:"status" db:"status"`
	InstallmentPlan  *InstallmentPlan       `json:"installment_plan,omitempty" db:"installment_plan"`
	Conditions       []AllocationCondition  `json:"conditions,omitempty" db:"conditions"`
	Milestones       []AllocationMilestone  `json:"milestones,omitempty" db:"milestones"`
	ApprovedAt       *time.Time             `json:"approved_at,omitempty" db:"approved_at"`
	DisbursedAt      *time.Time             `json:"disbursed_at,omitempty" db:"disbursed_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// InstallmentPlan represents a payment schedule for allocations
type InstallmentPlan struct {
	TotalInstallments int                   `json:"total_installments"`
	InstallmentAmount *Money                `json:"installment_amount"`
	Frequency         time.Duration         `json:"frequency"` // weekly, monthly, quarterly
	NextPaymentDate   time.Time             `json:"next_payment_date"`
	Installments      []InstallmentPayment  `json:"installments"`
}

// InstallmentPayment represents a single installment payment
type InstallmentPayment struct {
	InstallmentNumber int        `json:"installment_number"`
	Amount            *Money     `json:"amount"`
	ScheduledDate     time.Time  `json:"scheduled_date"`
	PaidDate          *time.Time `json:"paid_date,omitempty"`
	Status            string     `json:"status"` // "pending", "paid", "overdue"
	TransactionHash   string     `json:"transaction_hash,omitempty"`
}

// AllocationCondition represents conditions that must be met for allocation
type AllocationCondition struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"` // "milestone", "time", "performance", "approval"
	Description string      `json:"description"`
	Status      string      `json:"status"` // "pending", "met", "failed"
	Deadline    *time.Time  `json:"deadline,omitempty"`
	MetAt       *time.Time  `json:"met_at,omitempty"`
	Evidence    interface{} `json:"evidence,omitempty"`
}

// AllocationMilestone represents a milestone for allocation release
type AllocationMilestone struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Amount          *Money     `json:"amount"`
	Status          string     `json:"status"` // "pending", "completed", "verified"
	DueDate         time.Time  `json:"due_date"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
	VerificationRequired bool  `json:"verification_required"`
}

// GovernanceConfig represents DAO governance configuration
type GovernanceConfig struct {
	ID                    string                 `json:"id" db:"config_id"`
	Name                  string                 `json:"name" db:"name"`
	ProposalThreshold     *Money                 `json:"proposal_threshold" db:"proposal_threshold"`
	QuorumPercentage      float64                `json:"quorum_percentage" db:"quorum_percentage"`
	PassingThreshold      float64                `json:"passing_threshold" db:"passing_threshold"`
	VotingPeriod          time.Duration          `json:"voting_period" db:"voting_period"`
	ExecutionDelay        time.Duration          `json:"execution_delay" db:"execution_delay"`
	MaxActions            int                    `json:"max_actions" db:"max_actions"`
	TokenAddress          string                 `json:"token_address" db:"token_address"`
	TimelockAddress       string                 `json:"timelock_address" db:"timelock_address"`
	TreasuryAddress       string                 `json:"treasury_address" db:"treasury_address"`
	AllowDelegation       bool                   `json:"allow_delegation" db:"allow_delegation"`
	RequireReason         bool                   `json:"require_reason" db:"require_reason"`
	EmergencyPauseEnabled bool                   `json:"emergency_pause_enabled" db:"emergency_pause_enabled"`
	Parameters            map[string]interface{} `json:"parameters,omitempty" db:"parameters"`
	IsActive              bool                   `json:"is_active" db:"is_active"`
	CreatedAt             time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at" db:"updated_at"`
}

// VoteDelegation represents vote delegation between members
type VoteDelegation struct {
	ID           uuid.UUID              `json:"id" db:"delegation_id"`
	DelegatorID  uuid.UUID              `json:"delegator_id" db:"delegator_id"`
	DelegateID   uuid.UUID              `json:"delegate_id" db:"delegate_id"`
	ProposalType *ProposalType          `json:"proposal_type,omitempty" db:"proposal_type"` // nil for all proposals
	VotingPower  *Money                 `json:"voting_power" db:"voting_power"`
	IsActive     bool                   `json:"is_active" db:"is_active"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty" db:"expires_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// GovernanceEvent represents events in the governance system
type GovernanceEvent struct {
	ID         uuid.UUID              `json:"id" db:"event_id"`
	Type       string                 `json:"type" db:"type"` // "proposal_created", "vote_cast", "proposal_executed"
	ProposalID *uuid.UUID             `json:"proposal_id,omitempty" db:"proposal_id"`
	ActorID    uuid.UUID              `json:"actor_id" db:"actor_id"`
	Data       map[string]interface{} `json:"data,omitempty" db:"data"`
	TxHash     string                 `json:"tx_hash,omitempty" db:"tx_hash"`
	BlockHash  string                 `json:"block_hash,omitempty" db:"block_hash"`
	OccurredAt time.Time              `json:"occurred_at" db:"occurred_at"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// ProposalVoteResult represents aggregated voting results for a proposal
type ProposalVoteResult struct {
	ProposalID       uuid.UUID `json:"proposal_id"`
	TotalVotingPower *Money    `json:"total_voting_power"`
	VotesFor         *Money    `json:"votes_for"`
	VotesAgainst     *Money    `json:"votes_against"`
	VotesAbstain     *Money    `json:"votes_abstain"`
	VoterCount       int       `json:"voter_count"`
	QuorumReached    bool      `json:"quorum_reached"`
	Passed           bool      `json:"passed"`
	ParticipationRate float64  `json:"participation_rate"`
	ForPercentage    float64   `json:"for_percentage"`
	AgainstPercentage float64  `json:"against_percentage"`
	AbstainPercentage float64  `json:"abstain_percentage"`
}