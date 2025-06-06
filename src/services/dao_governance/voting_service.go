package dao_governance

import (
	"context"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// VotingServiceImpl implements the VotingService interface
type VotingServiceImpl struct {
	governanceRepo    repositories.GovernanceRepository
	blockchainService BlockchainIntegrationService
}

// NewVotingService creates a new voting service instance
func NewVotingService(
	governanceRepo repositories.GovernanceRepository,
	blockchainService BlockchainIntegrationService,
) *VotingServiceImpl {
	return &VotingServiceImpl{
		governanceRepo:    governanceRepo,
		blockchainService: blockchainService,
	}
}

// ValidateVoteEligibility checks if a voter is eligible to vote on a proposal
func (v *VotingServiceImpl) ValidateVoteEligibility(ctx context.Context, proposalID, voterID uuid.UUID) (*VoteEligibility, error) {
	// Get proposal
	proposal, err := v.governanceRepo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	// Check if proposal is in voting state
	if proposal.Status != entities.ProposalStatusActive {
		return &VoteEligibility{
			IsEligible: false,
			Reason:     fmt.Sprintf("proposal is not active (status: %s)", proposal.Status),
		}, nil
	}

	// Check if voting period is active
	now := time.Now()
	if now.Before(proposal.VotingStartTime) {
		return &VoteEligibility{
			IsEligible: false,
			Reason:     "voting has not started yet",
		}, nil
	}

	if now.After(proposal.VotingEndTime) {
		return &VoteEligibility{
			IsEligible: false,
			Reason:     "voting period has ended",
		}, nil
	}

	// Get voter
	voter, err := v.governanceRepo.GetMemberByID(ctx, voterID)
	if err != nil {
		return &VoteEligibility{
			IsEligible: false,
			Reason:     "voter not found",
		}, nil
	}

	// Check voter status
	if voter.Status != entities.MemberStatusActive {
		return &VoteEligibility{
			IsEligible: false,
			Reason:     fmt.Sprintf("voter is not active (status: %s)", voter.Status),
		}, nil
	}

	// Calculate voting power at proposal snapshot
	votingPower, err := v.CalculateVotingPower(ctx, voterID, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate voting power: %w", err)
	}

	// Check if voter has sufficient voting power
	if votingPower.Amount <= 0 {
		return &VoteEligibility{
			IsEligible: false,
			VotingPower: votingPower,
			Reason:     "no voting power at proposal snapshot",
		}, nil
	}

	// Check if voter has already voted
	existingVote, err := v.governanceRepo.GetVoteForProposalByVoter(ctx, proposalID, voterID)
	alreadyVoted := err == nil && existingVote != nil

	return &VoteEligibility{
		IsEligible:    true,
		VotingPower:   votingPower,
		AlreadyVoted:  alreadyVoted,
		BlockSnapshot: 0, // Would be set from proposal snapshot
	}, nil
}

// CalculateVotingPower calculates the voting power for a member at a specific proposal
func (v *VotingServiceImpl) CalculateVotingPower(ctx context.Context, memberID uuid.UUID, proposalID uuid.UUID) (*entities.Money, error) {
	// Get member
	member, err := v.governanceRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	// Get proposal for snapshot block
	proposal, err := v.governanceRepo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	// For now, use current voting power
	// In a full implementation, this would query the blockchain at the proposal snapshot block
	votingPower := member.VotingPower

	// Add delegated voting power
	delegatedPower, err := v.getDelegatedVotingPowerForProposal(ctx, memberID, proposalID)
	if err != nil {
		// Log error but continue with own voting power
		fmt.Printf("Failed to get delegated voting power: %v\n", err)
		delegatedPower = &entities.Money{Amount: 0, Currency: votingPower.Currency}
	}

	totalPower := &entities.Money{
		Amount:   votingPower.Amount + delegatedPower.Amount,
		Currency: votingPower.Currency,
	}

	return totalPower, nil
}

// ProcessVoteSubmission processes a vote submission
func (v *VotingServiceImpl) ProcessVoteSubmission(ctx context.Context, vote *entities.GovernanceVote) error {
	// Validate vote data
	if vote.ProposalID == uuid.Nil || vote.VoterID == uuid.Nil {
		return fmt.Errorf("invalid vote: missing proposal or voter ID")
	}

	if vote.Choice == "" {
		return fmt.Errorf("invalid vote: missing choice")
	}

	// Ensure vote has valid voting power
	if vote.VotingPower == nil || vote.VotingPower.Amount <= 0 {
		return fmt.Errorf("invalid vote: voting power must be positive")
	}

	// Set default weight if not provided
	if vote.Weight == 0 {
		vote.Weight = 1.0
	}

	// Validate weight is between 0 and 1
	if vote.Weight < 0 || vote.Weight > 1 {
		return fmt.Errorf("invalid vote: weight must be between 0 and 1")
	}

	// Set timestamps
	now := time.Now()
	vote.VotedAt = now
	vote.CreatedAt = now
	vote.UpdatedAt = now

	return nil
}

// TallyVotes tallies all votes for a proposal
func (v *VotingServiceImpl) TallyVotes(ctx context.Context, proposalID uuid.UUID) (*VoteTallyResult, error) {
	// Get all votes for the proposal
	votes, err := v.governanceRepo.GetVotesByProposal(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get votes: %w", err)
	}

	// Get proposal for quorum requirements
	proposal, err := v.governanceRepo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	// Initialize tallies
	result := &VoteTallyResult{
		ProposalID:       proposalID,
		TotalVotingPower: &entities.Money{Amount: 0, Currency: "TOKEN"},
		VotesFor:         &entities.Money{Amount: 0, Currency: "TOKEN"},
		VotesAgainst:     &entities.Money{Amount: 0, Currency: "TOKEN"},
		VotesAbstain:     &entities.Money{Amount: 0, Currency: "TOKEN"},
		VoterCount:       len(votes),
		PassingThreshold: proposal.PassingThreshold,
		TalliedAt:        time.Now(),
	}

	// Tally votes
	for _, vote := range votes {
		weightedPower := int64(float64(vote.VotingPower.Amount) * vote.Weight)
		
		switch vote.Choice {
		case entities.VoteChoiceFor:
			result.VotesFor.Amount += weightedPower
		case entities.VoteChoiceAgainst:
			result.VotesAgainst.Amount += weightedPower
		case entities.VoteChoiceAbstain:
			result.VotesAbstain.Amount += weightedPower
		}
		
		result.TotalVotingPower.Amount += weightedPower
	}

	// Calculate quorum and passing status
	quorumRequired, err := v.calculateQuorumRequired(ctx, proposal)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate quorum: %w", err)
	}

	result.QuorumRequired = quorumRequired
	result.QuorumReached = result.TotalVotingPower.Amount >= quorumRequired.Amount

	// Calculate passing status
	if result.QuorumReached {
		totalVotesForAgainst := result.VotesFor.Amount + result.VotesAgainst.Amount
		if totalVotesForAgainst > 0 {
			forPercentage := float64(result.VotesFor.Amount) / float64(totalVotesForAgainst)
			result.Passed = forPercentage >= proposal.PassingThreshold
		}
	}

	return result, nil
}

// CheckQuorumReached checks if quorum has been reached for a proposal
func (v *VotingServiceImpl) CheckQuorumReached(ctx context.Context, proposalID uuid.UUID) (bool, error) {
	tally, err := v.TallyVotes(ctx, proposalID)
	if err != nil {
		return false, fmt.Errorf("failed to tally votes: %w", err)
	}

	return tally.QuorumReached, nil
}

// DetermineProposalOutcome determines the final outcome of a proposal
func (v *VotingServiceImpl) DetermineProposalOutcome(ctx context.Context, proposalID uuid.UUID) (*ProposalOutcome, error) {
	// Get proposal
	proposal, err := v.governanceRepo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	// Tally votes
	voteResults, err := v.TallyVotes(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to tally votes: %w", err)
	}

	outcome := &ProposalOutcome{
		ProposalID:  proposalID,
		VoteResults: voteResults,
		CompletedAt: time.Now(),
	}

	// Determine final status based on voting results
	if !voteResults.QuorumReached {
		outcome.FinalStatus = entities.ProposalStatusRejected
		outcome.Reason = "Quorum not reached"
	} else if voteResults.Passed {
		outcome.FinalStatus = entities.ProposalStatusPassed
		outcome.Reason = "Proposal passed by vote"
	} else {
		outcome.FinalStatus = entities.ProposalStatusRejected
		outcome.Reason = "Proposal rejected by vote"
	}

	// Update proposal status
	proposal.Status = outcome.FinalStatus
	proposal.UpdatedAt = time.Now()

	if err := v.governanceRepo.UpdateProposal(ctx, proposal); err != nil {
		return nil, fmt.Errorf("failed to update proposal status: %w", err)
	}

	return outcome, nil
}

// Helper functions

// getDelegatedVotingPowerForProposal gets delegated voting power for a specific proposal type
func (v *VotingServiceImpl) getDelegatedVotingPowerForProposal(ctx context.Context, memberID uuid.UUID, proposalID uuid.UUID) (*entities.Money, error) {
	// Get proposal to determine type
	proposal, err := v.governanceRepo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	// Get all delegations where this member is the delegate
	delegations, err := v.governanceRepo.GetDelegationsByDelegate(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delegations: %w", err)
	}

	totalPower := &entities.Money{Amount: 0, Currency: "TOKEN"}
	
	for _, delegation := range delegations {
		// Check if delegation is active
		if !delegation.IsActive {
			continue
		}

		// Check if delegation has expired
		if delegation.ExpiresAt != nil && time.Now().After(*delegation.ExpiresAt) {
			continue
		}

		// Check if delegation applies to this proposal type
		if delegation.ProposalType != nil && *delegation.ProposalType != proposal.Type {
			continue
		}

		totalPower.Amount += delegation.VotingPower.Amount
	}

	return totalPower, nil
}

// calculateQuorumRequired calculates the required quorum for a proposal
func (v *VotingServiceImpl) calculateQuorumRequired(ctx context.Context, proposal *entities.GovernanceProposal) (*entities.Money, error) {
	// Get total voting power at the time of proposal creation
	// For now, we'll use a simplified calculation
	// In a full implementation, this would query the total supply at the proposal snapshot block
	
	// Get governance config for quorum percentage
	config, err := v.governanceRepo.GetActiveConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get governance config: %w", err)
	}

	// Get current total voting power (simplified)
	// This should be the total token supply at proposal snapshot
	distribution, err := v.governanceRepo.GetVotingPowerDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get voting power distribution: %w", err)
	}

	// Calculate required quorum
	quorumAmount := int64(float64(distribution.TotalSupply.Amount) * proposal.QuorumRequired)

	return &entities.Money{
		Amount:   quorumAmount,
		Currency: distribution.TotalSupply.Currency,
	}, nil
}

// ValidateVoteChoice validates that the vote choice is valid
func (v *VotingServiceImpl) ValidateVoteChoice(choice entities.VoteChoice) error {
	switch choice {
	case entities.VoteChoiceFor, entities.VoteChoiceAgainst, entities.VoteChoiceAbstain:
		return nil
	default:
		return fmt.Errorf("invalid vote choice: %s", choice)
	}
}

// CalculateVoteWeight calculates the weight of a vote based on various factors
func (v *VotingServiceImpl) CalculateVoteWeight(ctx context.Context, vote *entities.GovernanceVote) (float64, error) {
	// For now, return default weight of 1.0
	// In a more sophisticated system, this could factor in:
	// - Member reputation
	// - Proposal expertise matching
	// - Historical voting accuracy
	// - Stake duration
	
	return 1.0, nil
}

// GetVoteHistory returns the voting history for a member
func (v *VotingServiceImpl) GetVoteHistory(ctx context.Context, memberID uuid.UUID, limit int) ([]*entities.GovernanceVote, error) {
	filter := repositories.VoteFilter{
		Limit:     limit,
		SortBy:    "voted_at",
		SortOrder: "desc",
	}
	
	return v.governanceRepo.GetVotesByVoter(ctx, memberID, filter)
}

// GetActiveProposalsForVoting returns proposals that are currently in voting phase
func (v *VotingServiceImpl) GetActiveProposalsForVoting(ctx context.Context) ([]*entities.GovernanceProposal, error) {
	filter := repositories.ProposalFilter{
		Status: &entities.ProposalStatusActive,
		SortBy: "voting_end_time",
		SortOrder: "asc",
	}
	
	proposals, err := v.governanceRepo.ListProposals(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals: %w", err)
	}

	// Filter to only include proposals in active voting period
	var activeProposals []*entities.GovernanceProposal
	now := time.Now()
	
	for _, proposal := range proposals {
		if now.After(proposal.VotingStartTime) && now.Before(proposal.VotingEndTime) {
			activeProposals = append(activeProposals, proposal)
		}
	}

	return activeProposals, nil
}