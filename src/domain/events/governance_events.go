package events

import (
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// GovernanceEventType represents the type of governance event
type GovernanceEventType string

const (
	// Proposal Events
	GovernanceEventProposalCreated   GovernanceEventType = "governance.proposal.created"
	GovernanceEventProposalSubmitted GovernanceEventType = "governance.proposal.submitted"
	GovernanceEventProposalStarted   GovernanceEventType = "governance.proposal.started"
	GovernanceEventProposalEnded     GovernanceEventType = "governance.proposal.ended"
	GovernanceEventProposalExecuted  GovernanceEventType = "governance.proposal.executed"
	GovernanceEventProposalCanceled  GovernanceEventType = "governance.proposal.canceled"

	// Voting Events
	GovernanceEventVoteCast       GovernanceEventType = "governance.vote.cast"
	GovernanceEventVoteChanged    GovernanceEventType = "governance.vote.changed"
	GovernanceEventVoteDelegated  GovernanceEventType = "governance.vote.delegated"
	GovernanceEventDelegationRevoked GovernanceEventType = "governance.delegation.revoked"

	// Member Events
	GovernanceEventMemberJoined   GovernanceEventType = "governance.member.joined"
	GovernanceEventMemberUpdated  GovernanceEventType = "governance.member.updated"
	GovernanceEventMemberSuspended GovernanceEventType = "governance.member.suspended"
	GovernanceEventMemberRemoved  GovernanceEventType = "governance.member.removed"

	// Treasury Events
	GovernanceEventAllocationCreated  GovernanceEventType = "governance.allocation.created"
	GovernanceEventAllocationApproved GovernanceEventType = "governance.allocation.approved"
	GovernanceEventAllocationDisbursed GovernanceEventType = "governance.allocation.disbursed"
	GovernanceEventAllocationCompleted GovernanceEventType = "governance.allocation.completed"

	// Configuration Events
	GovernanceEventConfigUpdated      GovernanceEventType = "governance.config.updated"
	GovernanceEventEmergencyActivated GovernanceEventType = "governance.emergency.activated"
	GovernanceEventEmergencyResolved  GovernanceEventType = "governance.emergency.resolved"
)

// ProposalCreatedEvent represents a new proposal being created
type ProposalCreatedEvent struct {
	BaseEvent
	ProposalID    uuid.UUID                  `json:"proposal_id"`
	Title         string                     `json:"title"`
	Type          entities.ProposalType      `json:"type"`
	ProposerID    uuid.UUID                  `json:"proposer_id"`
	Description   string                     `json:"description"`
	VotingPower   *entities.Money            `json:"voting_power"`
	Parameters    map[string]interface{}     `json:"parameters,omitempty"`
	Actions       []entities.ProposalAction  `json:"actions,omitempty"`
}

// ProposalSubmittedEvent represents a proposal being submitted for voting
type ProposalSubmittedEvent struct {
	BaseEvent
	ProposalID       uuid.UUID           `json:"proposal_id"`
	OnChainProposalID string             `json:"on_chain_proposal_id"`
	VotingStartTime  time.Time           `json:"voting_start_time"`
	VotingEndTime    time.Time           `json:"voting_end_time"`
	QuorumRequired   float64             `json:"quorum_required"`
	PassingThreshold float64             `json:"passing_threshold"`
	TxHash           string              `json:"tx_hash"`
}

// ProposalExecutedEvent represents a proposal being executed
type ProposalExecutedEvent struct {
	BaseEvent
	ProposalID      uuid.UUID                  `json:"proposal_id"`
	ExecutorID      uuid.UUID                  `json:"executor_id"`
	ActionsExecuted []entities.ProposalAction  `json:"actions_executed"`
	Results         map[string]interface{}     `json:"results"`
	TxHashes        []string                   `json:"tx_hashes"`
	Success         bool                       `json:"success"`
	ErrorMessage    string                     `json:"error_message,omitempty"`
}

// VoteCastEvent represents a vote being cast on a proposal
type VoteCastEvent struct {
	BaseEvent
	VoteID           uuid.UUID           `json:"vote_id"`
	ProposalID       uuid.UUID           `json:"proposal_id"`
	VoterID          uuid.UUID           `json:"voter_id"`
	Choice           entities.VoteChoice `json:"choice"`
	VotingPower      *entities.Money     `json:"voting_power"`
	Weight           float64             `json:"weight"`
	DelegatedFrom    []uuid.UUID         `json:"delegated_from,omitempty"`
	Rationale        string              `json:"rationale,omitempty"`
	OnChainTxHash    string              `json:"on_chain_tx_hash,omitempty"`
}

// VoteDelegatedEvent represents voting power being delegated
type VoteDelegatedEvent struct {
	BaseEvent
	DelegationID  uuid.UUID                 `json:"delegation_id"`
	DelegatorID   uuid.UUID                 `json:"delegator_id"`
	DelegateID    uuid.UUID                 `json:"delegate_id"`
	ProposalType  *entities.ProposalType    `json:"proposal_type,omitempty"`
	VotingPower   *entities.Money           `json:"voting_power"`
	ExpiresAt     *time.Time                `json:"expires_at,omitempty"`
}

// MemberJoinedEvent represents a new member joining the DAO
type MemberJoinedEvent struct {
	BaseEvent
	MemberID          uuid.UUID               `json:"member_id"`
	Address           string                  `json:"address"`
	Role              entities.MemberRole     `json:"role"`
	TokenBalance      *entities.Money         `json:"token_balance"`
	VotingPower       *entities.Money         `json:"voting_power"`
	ContributionScore float64                 `json:"contribution_score"`
	VestingSchedule   *entities.VestingSchedule `json:"vesting_schedule,omitempty"`
}

// MemberUpdatedEvent represents a member's information being updated
type MemberUpdatedEvent struct {
	BaseEvent
	MemberID          uuid.UUID               `json:"member_id"`
	Changes           map[string]interface{}  `json:"changes"`
	PreviousRole      entities.MemberRole     `json:"previous_role,omitempty"`
	NewRole           entities.MemberRole     `json:"new_role,omitempty"`
	UpdatedBy         uuid.UUID               `json:"updated_by"`
}

// AllocationCreatedEvent represents a treasury allocation being created
type AllocationCreatedEvent struct {
	BaseEvent
	AllocationID     uuid.UUID                      `json:"allocation_id"`
	ProposalID       uuid.UUID                      `json:"proposal_id"`
	Title            string                         `json:"title"`
	Amount           *entities.Money                `json:"amount"`
	RecipientID      uuid.UUID                      `json:"recipient_id"`
	Category         string                         `json:"category"`
	InstallmentPlan  *entities.InstallmentPlan      `json:"installment_plan,omitempty"`
	Conditions       []entities.AllocationCondition `json:"conditions,omitempty"`
	Milestones       []entities.AllocationMilestone `json:"milestones,omitempty"`
}

// AllocationDisbursedEvent represents funds being disbursed from allocation
type AllocationDisbursedEvent struct {
	BaseEvent
	AllocationID      uuid.UUID       `json:"allocation_id"`
	Amount            *entities.Money `json:"amount"`
	RecipientAddress  string          `json:"recipient_address"`
	InstallmentNumber *int            `json:"installment_number,omitempty"`
	MilestoneID       *string         `json:"milestone_id,omitempty"`
	TxHash            string          `json:"tx_hash"`
	DisbursedBy       uuid.UUID       `json:"disbursed_by"`
}

// ConfigUpdatedEvent represents governance configuration being updated
type ConfigUpdatedEvent struct {
	BaseEvent
	ConfigID         string                 `json:"config_id"`
	Changes          map[string]interface{} `json:"changes"`
	PreviousConfig   entities.GovernanceConfig `json:"previous_config"`
	NewConfig        entities.GovernanceConfig `json:"new_config"`
	UpdatedBy        uuid.UUID              `json:"updated_by"`
	EffectiveDate    time.Time              `json:"effective_date"`
}

// EmergencyActivatedEvent represents emergency mode being activated
type EmergencyActivatedEvent struct {
	BaseEvent
	Reason           string                 `json:"reason"`
	TriggerCondition string                 `json:"trigger_condition"`
	ActivatedBy      uuid.UUID              `json:"activated_by"`
	Severity         entities.RiskSeverity  `json:"severity"`
	Actions          []string               `json:"actions"`
	ExpectedDuration time.Duration          `json:"expected_duration,omitempty"`
}

// EmergencyResolvedEvent represents emergency mode being resolved
type EmergencyResolvedEvent struct {
	BaseEvent
	Duration         time.Duration          `json:"duration"`
	Resolution       string                 `json:"resolution"`
	ActionsCompleted []string               `json:"actions_completed"`
	ResolvedBy       uuid.UUID              `json:"resolved_by"`
	LessonsLearned   string                 `json:"lessons_learned,omitempty"`
}

// GovernanceMetricsUpdatedEvent represents governance metrics being updated
type GovernanceMetricsUpdatedEvent struct {
	BaseEvent
	TotalMembers         int                    `json:"total_members"`
	ActiveMembers        int                    `json:"active_members"`
	TotalProposals       int                    `json:"total_proposals"`
	ActiveProposals      int                    `json:"active_proposals"`
	TotalVotingPower     *entities.Money        `json:"total_voting_power"`
	ParticipationRate    float64                `json:"participation_rate"`
	TreasuryBalance      *entities.Money        `json:"treasury_balance"`
	PendingAllocations   *entities.Money        `json:"pending_allocations"`
	MetricsUpdatedAt     time.Time              `json:"metrics_updated_at"`
}

// Helper functions for creating governance events

// NewProposalCreatedEvent creates a new proposal created event
func NewProposalCreatedEvent(proposal *entities.GovernanceProposal) *ProposalCreatedEvent {
	return &ProposalCreatedEvent{
		BaseEvent: BaseEvent{
			EventID:   uuid.New(),
			EventType: string(GovernanceEventProposalCreated),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "governance-service",
			},
		},
		ProposalID:  proposal.ID,
		Title:       proposal.Title,
		Type:        proposal.Type,
		ProposerID:  proposal.ProposerID,
		Description: proposal.Description,
		VotingPower: proposal.VotingPower,
		Parameters:  proposal.Parameters,
		Actions:     proposal.Actions,
	}
}

// NewVoteCastEvent creates a new vote cast event
func NewVoteCastEvent(vote *entities.GovernanceVote) *VoteCastEvent {
	return &VoteCastEvent{
		BaseEvent: BaseEvent{
			EventID:   uuid.New(),
			EventType: string(GovernanceEventVoteCast),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "governance-service",
			},
		},
		VoteID:        vote.ID,
		ProposalID:    vote.ProposalID,
		VoterID:       vote.VoterID,
		Choice:        vote.Choice,
		VotingPower:   vote.VotingPower,
		Weight:        vote.Weight,
		DelegatedFrom: vote.DelegatedFrom,
		Rationale:     vote.Rationale,
		OnChainTxHash: vote.OnChainTxHash,
	}
}

// NewMemberJoinedEvent creates a new member joined event
func NewMemberJoinedEvent(member *entities.DAOMember) *MemberJoinedEvent {
	return &MemberJoinedEvent{
		BaseEvent: BaseEvent{
			EventID:   uuid.New(),
			EventType: string(GovernanceEventMemberJoined),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "governance-service",
			},
		},
		MemberID:          member.ID,
		Address:           member.Address,
		Role:              member.Role,
		TokenBalance:      member.TokenBalance,
		VotingPower:       member.VotingPower,
		ContributionScore: member.ContributionScore,
		VestingSchedule:   member.VestingSchedule,
	}
}

// NewAllocationDisbursedEvent creates a new allocation disbursed event
func NewAllocationDisbursedEvent(allocation *entities.TreasuryAllocation, amount *entities.Money, txHash string, disbursedBy uuid.UUID) *AllocationDisbursedEvent {
	return &AllocationDisbursedEvent{
		BaseEvent: BaseEvent{
			EventID:   uuid.New(),
			EventType: string(GovernanceEventAllocationDisbursed),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"source": "governance-service",
			},
		},
		AllocationID:     allocation.ID,
		Amount:           amount,
		RecipientAddress: allocation.RecipientAddress,
		TxHash:           txHash,
		DisbursedBy:      disbursedBy,
	}
}