package repositories

import (
	"context"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/google/uuid"
)

// GovernanceRepository provides data access for governance operations
type GovernanceRepository interface {
	// Proposal Management
	CreateProposal(ctx context.Context, proposal *entities.GovernanceProposal) error
	GetProposalByID(ctx context.Context, id uuid.UUID) (*entities.GovernanceProposal, error)
	UpdateProposal(ctx context.Context, proposal *entities.GovernanceProposal) error
	ListProposals(ctx context.Context, filter ProposalFilter) ([]*entities.GovernanceProposal, error)
	CountProposals(ctx context.Context, filter ProposalFilter) (int, error)
	GetActiveProposals(ctx context.Context) ([]*entities.GovernanceProposal, error)
	GetProposalsByStatus(ctx context.Context, status entities.ProposalStatus) ([]*entities.GovernanceProposal, error)
	GetProposalsByType(ctx context.Context, proposalType entities.ProposalType) ([]*entities.GovernanceProposal, error)
	GetProposalsByProposer(ctx context.Context, proposerID uuid.UUID) ([]*entities.GovernanceProposal, error)
	GetExpiredProposals(ctx context.Context) ([]*entities.GovernanceProposal, error)
	DeleteProposal(ctx context.Context, id uuid.UUID) error

	// Vote Management
	CreateVote(ctx context.Context, vote *entities.GovernanceVote) error
	GetVoteByID(ctx context.Context, id uuid.UUID) (*entities.GovernanceVote, error)
	UpdateVote(ctx context.Context, vote *entities.GovernanceVote) error
	GetVotesByProposal(ctx context.Context, proposalID uuid.UUID) ([]*entities.GovernanceVote, error)
	GetVotesByVoter(ctx context.Context, voterID uuid.UUID, filter VoteFilter) ([]*entities.GovernanceVote, error)
	GetVoteForProposalByVoter(ctx context.Context, proposalID, voterID uuid.UUID) (*entities.GovernanceVote, error)
	CountVotesByProposal(ctx context.Context, proposalID uuid.UUID) (int, error)
	GetVoteResults(ctx context.Context, proposalID uuid.UUID) (*entities.ProposalVoteResult, error)
	DeleteVote(ctx context.Context, id uuid.UUID) error

	// Member Management
	CreateMember(ctx context.Context, member *entities.DAOMember) error
	GetMemberByID(ctx context.Context, id uuid.UUID) (*entities.DAOMember, error)
	GetMemberByAddress(ctx context.Context, address string) (*entities.DAOMember, error)
	UpdateMember(ctx context.Context, member *entities.DAOMember) error
	ListMembers(ctx context.Context, filter MemberFilter) ([]*entities.DAOMember, error)
	CountMembers(ctx context.Context, filter MemberFilter) (int, error)
	GetMembersByRole(ctx context.Context, role entities.MemberRole) ([]*entities.DAOMember, error)
	GetActiveMembers(ctx context.Context) ([]*entities.DAOMember, error)
	GetTopMembersByContribution(ctx context.Context, limit int) ([]*entities.DAOMember, error)
	UpdateMemberVotingPower(ctx context.Context, memberID uuid.UUID, votingPower *entities.Money) error
	DeleteMember(ctx context.Context, id uuid.UUID) error

	// Treasury Allocation Management
	CreateAllocation(ctx context.Context, allocation *entities.TreasuryAllocation) error
	GetAllocationByID(ctx context.Context, id uuid.UUID) (*entities.TreasuryAllocation, error)
	UpdateAllocation(ctx context.Context, allocation *entities.TreasuryAllocation) error
	ListAllocations(ctx context.Context, filter AllocationFilter) ([]*entities.TreasuryAllocation, error)
	CountAllocations(ctx context.Context, filter AllocationFilter) (int, error)
	GetAllocationsByProposal(ctx context.Context, proposalID uuid.UUID) ([]*entities.TreasuryAllocation, error)
	GetAllocationsByRecipient(ctx context.Context, recipientID uuid.UUID) ([]*entities.TreasuryAllocation, error)
	GetAllocationsByStatus(ctx context.Context, status entities.AllocationStatus) ([]*entities.TreasuryAllocation, error)
	GetPendingAllocations(ctx context.Context) ([]*entities.TreasuryAllocation, error)
	GetExpiredAllocations(ctx context.Context) ([]*entities.TreasuryAllocation, error)
	DeleteAllocation(ctx context.Context, id uuid.UUID) error

	// Vote Delegation Management
	CreateDelegation(ctx context.Context, delegation *entities.VoteDelegation) error
	GetDelegationByID(ctx context.Context, id uuid.UUID) (*entities.VoteDelegation, error)
	UpdateDelegation(ctx context.Context, delegation *entities.VoteDelegation) error
	GetDelegationsByDelegator(ctx context.Context, delegatorID uuid.UUID) ([]*entities.VoteDelegation, error)
	GetDelegationsByDelegate(ctx context.Context, delegateID uuid.UUID) ([]*entities.VoteDelegation, error)
	GetActiveDelegations(ctx context.Context) ([]*entities.VoteDelegation, error)
	RevokeDelegation(ctx context.Context, id uuid.UUID) error
	DeleteDelegation(ctx context.Context, id uuid.UUID) error

	// Configuration Management
	CreateConfig(ctx context.Context, config *entities.GovernanceConfig) error
	GetConfigByID(ctx context.Context, id string) (*entities.GovernanceConfig, error)
	GetActiveConfig(ctx context.Context) (*entities.GovernanceConfig, error)
	UpdateConfig(ctx context.Context, config *entities.GovernanceConfig) error
	ListConfigs(ctx context.Context) ([]*entities.GovernanceConfig, error)
	DeleteConfig(ctx context.Context, id string) error

	// Event Management
	CreateEvent(ctx context.Context, event *entities.GovernanceEvent) error
	GetEventsByProposal(ctx context.Context, proposalID uuid.UUID) ([]*entities.GovernanceEvent, error)
	GetEventsByActor(ctx context.Context, actorID uuid.UUID, limit int) ([]*entities.GovernanceEvent, error)
	GetEventsByType(ctx context.Context, eventType string, filter EventFilter) ([]*entities.GovernanceEvent, error)
	GetRecentEvents(ctx context.Context, limit int) ([]*entities.GovernanceEvent, error)

	// Analytics and Statistics
	GetGovernanceMetrics(ctx context.Context, timeRange TimeRange) (*GovernanceMetrics, error)
	GetMemberParticipationStats(ctx context.Context, memberID uuid.UUID, timeRange TimeRange) (*MemberParticipationStats, error)
	GetProposalSuccessRate(ctx context.Context, timeRange TimeRange) (float64, error)
	GetVotingPowerDistribution(ctx context.Context) (*VotingPowerDistribution, error)
	GetTreasuryAllocationSummary(ctx context.Context, timeRange TimeRange) (*TreasuryAllocationSummary, error)

	// Batch Operations
	CreateProposalsInBatch(ctx context.Context, proposals []*entities.GovernanceProposal) error
	CreateVotesInBatch(ctx context.Context, votes []*entities.GovernanceVote) error
	UpdateMembersInBatch(ctx context.Context, members []*entities.DAOMember) error
}

// Filter types for repository queries

// ProposalFilter represents filtering options for proposals
type ProposalFilter struct {
	Status         *entities.ProposalStatus `json:"status,omitempty"`
	Type           *entities.ProposalType   `json:"type,omitempty"`
	ProposerID     *uuid.UUID               `json:"proposer_id,omitempty"`
	StartTime      *time.Time               `json:"start_time,omitempty"`
	EndTime        *time.Time               `json:"end_time,omitempty"`
	MinVotingPower *entities.Money          `json:"min_voting_power,omitempty"`
	MaxVotingPower *entities.Money          `json:"max_voting_power,omitempty"`
	SearchText     string                   `json:"search_text,omitempty"`
	Limit          int                      `json:"limit,omitempty"`
	Offset         int                      `json:"offset,omitempty"`
	SortBy         string                   `json:"sort_by,omitempty"`
	SortOrder      string                   `json:"sort_order,omitempty"`
}

// VoteFilter represents filtering options for votes
type VoteFilter struct {
	ProposalID     *uuid.UUID           `json:"proposal_id,omitempty"`
	Choice         *entities.VoteChoice `json:"choice,omitempty"`
	MinVotingPower *entities.Money      `json:"min_voting_power,omitempty"`
	MaxVotingPower *entities.Money      `json:"max_voting_power,omitempty"`
	StartTime      *time.Time           `json:"start_time,omitempty"`
	EndTime        *time.Time           `json:"end_time,omitempty"`
	HasRationale   *bool                `json:"has_rationale,omitempty"`
	Limit          int                  `json:"limit,omitempty"`
	Offset         int                  `json:"offset,omitempty"`
	SortBy         string               `json:"sort_by,omitempty"`
	SortOrder      string               `json:"sort_order,omitempty"`
}

// MemberFilter represents filtering options for members
type MemberFilter struct {
	Role               *entities.MemberRole   `json:"role,omitempty"`
	Status             *entities.MemberStatus `json:"status,omitempty"`
	MinTokenBalance    *entities.Money        `json:"min_token_balance,omitempty"`
	MaxTokenBalance    *entities.Money        `json:"max_token_balance,omitempty"`
	MinVotingPower     *entities.Money        `json:"min_voting_power,omitempty"`
	MaxVotingPower     *entities.Money        `json:"max_voting_power,omitempty"`
	MinContribution    *float64               `json:"min_contribution,omitempty"`
	MaxContribution    *float64               `json:"max_contribution,omitempty"`
	JoinedAfter        *time.Time             `json:"joined_after,omitempty"`
	JoinedBefore       *time.Time             `json:"joined_before,omitempty"`
	LastActivityAfter  *time.Time             `json:"last_activity_after,omitempty"`
	LastActivityBefore *time.Time             `json:"last_activity_before,omitempty"`
	HasDelegatedTo     *uuid.UUID             `json:"has_delegated_to,omitempty"`
	SearchText         string                 `json:"search_text,omitempty"`
	Limit              int                    `json:"limit,omitempty"`
	Offset             int                    `json:"offset,omitempty"`
	SortBy             string                 `json:"sort_by,omitempty"`
	SortOrder          string                 `json:"sort_order,omitempty"`
}

// AllocationFilter represents filtering options for allocations
type AllocationFilter struct {
	ProposalID     *uuid.UUID                 `json:"proposal_id,omitempty"`
	RecipientID    *uuid.UUID                 `json:"recipient_id,omitempty"`
	Status         *entities.AllocationStatus `json:"status,omitempty"`
	Category       *string                    `json:"category,omitempty"`
	MinAmount      *entities.Money            `json:"min_amount,omitempty"`
	MaxAmount      *entities.Money            `json:"max_amount,omitempty"`
	Currency       *string                    `json:"currency,omitempty"`
	ApprovedAfter  *time.Time                 `json:"approved_after,omitempty"`
	ApprovedBefore *time.Time                 `json:"approved_before,omitempty"`
	SearchText     string                     `json:"search_text,omitempty"`
	Limit          int                        `json:"limit,omitempty"`
	Offset         int                        `json:"offset,omitempty"`
	SortBy         string                     `json:"sort_by,omitempty"`
	SortOrder      string                     `json:"sort_order,omitempty"`
}

// EventFilter represents filtering options for events
type EventFilter struct {
	ActorID   *uuid.UUID `json:"actor_id,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Analytics result types

// GovernanceMetrics represents governance system metrics
type GovernanceMetrics struct {
	TotalMembers         int             `json:"total_members"`
	ActiveMembers        int             `json:"active_members"`
	TotalProposals       int             `json:"total_proposals"`
	ActiveProposals      int             `json:"active_proposals"`
	PassedProposals      int             `json:"passed_proposals"`
	RejectedProposals    int             `json:"rejected_proposals"`
	TotalVotes           int             `json:"total_votes"`
	TotalVotingPower     *entities.Money `json:"total_voting_power"`
	ParticipationRate    float64         `json:"participation_rate"`
	AverageQuorum        float64         `json:"average_quorum"`
	TreasuryBalance      *entities.Money `json:"treasury_balance"`
	TotalAllocations     *entities.Money `json:"total_allocations"`
	PendingAllocations   *entities.Money `json:"pending_allocations"`
	CompletedAllocations *entities.Money `json:"completed_allocations"`
	ProposalSuccessRate  float64         `json:"proposal_success_rate"`
	AverageVotingPeriod  time.Duration   `json:"average_voting_period"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// MemberParticipationStats represents participation statistics for a member
type MemberParticipationStats struct {
	MemberID                uuid.UUID       `json:"member_id"`
	ProposalsSubmitted      int             `json:"proposals_submitted"`
	ProposalSuccessRate     float64         `json:"proposal_success_rate"`
	VotesParticipated       int             `json:"votes_participated"`
	VotingParticipationRate float64         `json:"voting_participation_rate"`
	TotalVotingPower        *entities.Money `json:"total_voting_power"`
	VotesFor                int             `json:"votes_for"`
	VotesAgainst            int             `json:"votes_against"`
	VotesAbstain            int             `json:"votes_abstain"`
	DelegationsReceived     int             `json:"delegations_received"`
	DelegationsMade         int             `json:"delegations_made"`
	LastActivity            time.Time       `json:"last_activity"`
	ContributionScore       float64         `json:"contribution_score"`
	Period                  TimeRange       `json:"period"`
}

// VotingPowerDistribution represents the distribution of voting power
type VotingPowerDistribution struct {
	TotalSupply        *entities.Money     `json:"total_supply"`
	CirculatingSupply  *entities.Money     `json:"circulating_supply"`
	StakedSupply       *entities.Money     `json:"staked_supply"`
	DelegatedSupply    *entities.Money     `json:"delegated_supply"`
	TopHolders         []MemberVotingPower `json:"top_holders"`
	ConcentrationRatio float64             `json:"concentration_ratio"` // Percentage held by top 10 holders
	HHI                float64             `json:"hhi"`                 // Herfindahl-Hirschman Index
	GiniCoefficient    float64             `json:"gini_coefficient"`
	MedianHolding      *entities.Money     `json:"median_holding"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

// MemberVotingPower represents a member's voting power
type MemberVotingPower struct {
	MemberID    uuid.UUID       `json:"member_id"`
	Address     string          `json:"address"`
	VotingPower *entities.Money `json:"voting_power"`
	Percentage  float64         `json:"percentage"`
}

// TreasuryAllocationSummary represents summary of treasury allocations
type TreasuryAllocationSummary struct {
	TotalAllocated        *entities.Money            `json:"total_allocated"`
	TotalDisbursed        *entities.Money            `json:"total_disbursed"`
	TotalPending          *entities.Money            `json:"total_pending"`
	AllocationsByCategory map[string]*entities.Money `json:"allocations_by_category"`
	AllocationsByStatus   map[string]*entities.Money `json:"allocations_by_status"`
	AverageAllocationSize *entities.Money            `json:"average_allocation_size"`
	TotalRecipients       int                        `json:"total_recipients"`
	CompletionRate        float64                    `json:"completion_rate"`
	AverageCompletion     time.Duration              `json:"average_completion"`
	Period                TimeRange                  `json:"period"`
	UpdatedAt             time.Time                  `json:"updated_at"`
}
