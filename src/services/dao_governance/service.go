package dao_governance

import (
	"context"
	"fmt"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/events"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// Service implements the GovernanceService interface
type Service struct {
	governanceRepo   repositories.GovernanceRepository
	eventRepo        repositories.EventRepository
	votingService    VotingService
	membershipService MembershipService
	treasuryService  TreasuryGovernanceService
	blockchainService BlockchainIntegrationService
	orchestratorService ProposalOrchestratorService
}

// NewService creates a new governance service instance
func NewService(
	governanceRepo repositories.GovernanceRepository,
	eventRepo repositories.EventRepository,
	votingService VotingService,
	membershipService MembershipService,
	treasuryService TreasuryGovernanceService,
	blockchainService BlockchainIntegrationService,
	orchestratorService ProposalOrchestratorService,
) *Service {
	return &Service{
		governanceRepo:      governanceRepo,
		eventRepo:           eventRepo,
		votingService:       votingService,
		membershipService:   membershipService,
		treasuryService:     treasuryService,
		blockchainService:   blockchainService,
		orchestratorService: orchestratorService,
	}
}

// Proposal Management

func (s *Service) CreateProposal(ctx context.Context, request ProposalCreationRequest) (*entities.GovernanceProposal, error) {
	// Validate proposer exists and has sufficient voting power
	member, err := s.governanceRepo.GetMemberByID(ctx, request.ProposerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposer: %w", err)
	}

	// Check if member has sufficient voting power for proposal type
	config, err := s.GetGovernanceConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get governance config: %w", err)
	}

	if member.VotingPower.Amount < config.ProposalThreshold.Amount {
		return nil, fmt.Errorf("insufficient voting power: required %.2f, has %.2f", 
			config.ProposalThreshold.Amount, member.VotingPower.Amount)
	}

	// Set default values based on governance config and proposal type
	proposal := &entities.GovernanceProposal{
		ID:                uuid.New(),
		Title:             request.Title,
		Description:       request.Description,
		Type:              request.Type,
		Status:            entities.ProposalStatusDraft,
		ProposerID:        request.ProposerID,
		ProposerAddress:   member.Address,
		VotingPower:       member.VotingPower,
		Actions:           request.Actions,
		Parameters:        request.Parameters,
		Metadata:          make(map[string]interface{}),
		IPFSHash:          request.IPFSHash,
		SubmittedAt:       time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Set proposal-specific parameters
	if request.QuorumRequired != nil {
		proposal.QuorumRequired = *request.QuorumRequired
	} else {
		proposal.QuorumRequired = config.QuorumPercentage / 100
	}

	if request.PassingThreshold != nil {
		proposal.PassingThreshold = *request.PassingThreshold
	} else {
		proposal.PassingThreshold = config.PassingThreshold
	}

	if request.VotingPeriod != nil {
		proposal.VotingEndTime = time.Now().Add(*request.VotingPeriod)
	} else {
		proposal.VotingEndTime = time.Now().Add(config.VotingPeriod)
	}

	if request.ExecutionDelay != nil {
		proposal.ExecutionDelay = *request.ExecutionDelay
	} else {
		proposal.ExecutionDelay = config.ExecutionDelay
	}

	// Add metadata
	if request.Category != "" {
		proposal.Metadata["category"] = request.Category
	}
	if request.DiscussionURL != "" {
		proposal.Metadata["discussion_url"] = request.DiscussionURL
	}
	if request.IsEmergency {
		proposal.Metadata["is_emergency"] = true
	}

	// Create proposal in database
	if err := s.governanceRepo.CreateProposal(ctx, proposal); err != nil {
		return nil, fmt.Errorf("failed to create proposal: %w", err)
	}

	// Create proposal created event
	event := events.NewProposalCreatedEvent(proposal)
	if err := s.eventRepo.Save(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create proposal created event: %v\n", err)
	}

	return proposal, nil
}

func (s *Service) SubmitProposal(ctx context.Context, proposalID uuid.UUID) error {
	proposal, err := s.governanceRepo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal: %w", err)
	}

	if proposal.Status != entities.ProposalStatusDraft {
		return fmt.Errorf("proposal must be in draft status to submit")
	}

	// Submit proposal on-chain
	onChainID, err := s.blockchainService.SubmitProposalOnChain(ctx, proposal)
	if err != nil {
		return fmt.Errorf("failed to submit proposal on-chain: %w", err)
	}

	// Update proposal status and on-chain ID
	proposal.Status = entities.ProposalStatusSubmitted
	proposal.OnChainProposalID = &onChainID
	proposal.VotingStartTime = time.Now()
	proposal.UpdatedAt = time.Now()

	if err := s.governanceRepo.UpdateProposal(ctx, proposal); err != nil {
		return fmt.Errorf("failed to update proposal: %w", err)
	}

	// Start proposal workflow
	workflowRequest := ProposalWorkflowRequest{
		ProposalID:    proposalID,
		WorkflowType:  "standard",
		AutoAdvance:   true,
		Configuration: map[string]interface{}{
			"voting_period": proposal.VotingEndTime.Sub(proposal.VotingStartTime).String(),
			"quorum_required": proposal.QuorumRequired,
		},
	}

	_, err = s.orchestratorService.StartProposalWorkflow(ctx, workflowRequest)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to start proposal workflow: %v\n", err)
	}

	return nil
}

func (s *Service) GetProposal(ctx context.Context, proposalID uuid.UUID) (*entities.GovernanceProposal, error) {
	return s.governanceRepo.GetProposalByID(ctx, proposalID)
}

func (s *Service) ListProposals(ctx context.Context, filter repositories.ProposalFilter) ([]*entities.GovernanceProposal, error) {
	return s.governanceRepo.ListProposals(ctx, filter)
}

func (s *Service) UpdateProposal(ctx context.Context, proposalID uuid.UUID, updates ProposalUpdates) error {
	proposal, err := s.governanceRepo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal: %w", err)
	}

	if proposal.Status != entities.ProposalStatusDraft {
		return fmt.Errorf("can only update draft proposals")
	}

	// Apply updates
	if updates.Title != nil {
		proposal.Title = *updates.Title
	}
	if updates.Description != nil {
		proposal.Description = *updates.Description
	}
	if updates.Actions != nil {
		proposal.Actions = *updates.Actions
	}
	if updates.Parameters != nil {
		proposal.Parameters = updates.Parameters
	}
	if updates.IPFSHash != nil {
		proposal.IPFSHash = *updates.IPFSHash
	}
	if updates.DiscussionURL != nil {
		proposal.Metadata["discussion_url"] = *updates.DiscussionURL
	}

	proposal.UpdatedAt = time.Now()

	return s.governanceRepo.UpdateProposal(ctx, proposal)
}

func (s *Service) CancelProposal(ctx context.Context, proposalID uuid.UUID, reason string) error {
	proposal, err := s.governanceRepo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal: %w", err)
	}

	if proposal.Status == entities.ProposalStatusExecuted || proposal.Status == entities.ProposalStatusCanceled {
		return fmt.Errorf("cannot cancel proposal in status: %s", proposal.Status)
	}

	proposal.Status = entities.ProposalStatusCanceled
	proposal.UpdatedAt = time.Now()
	proposal.Metadata["cancellation_reason"] = reason

	return s.governanceRepo.UpdateProposal(ctx, proposal)
}

func (s *Service) ExecuteProposal(ctx context.Context, proposalID uuid.UUID) (*ProposalExecutionResult, error) {
	proposal, err := s.governanceRepo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	if proposal.Status != entities.ProposalStatusPassed {
		return nil, fmt.Errorf("proposal must be passed to execute")
	}

	// Check if execution delay has passed
	if proposal.ExecutionDeadline != nil && time.Now().Before(*proposal.ExecutionDeadline) {
		return nil, fmt.Errorf("execution delay not yet passed")
	}

	// Execute proposal on-chain
	txHashes, err := s.blockchainService.ExecuteProposalOnChain(ctx, proposalID)
	if err != nil {
		return &ProposalExecutionResult{
			ProposalID:   proposalID,
			Success:      false,
			ErrorMessage: err.Error(),
			ExecutedAt:   time.Now(),
		}, fmt.Errorf("failed to execute proposal on-chain: %w", err)
	}

	// Update proposal status
	proposal.Status = entities.ProposalStatusExecuted
	proposal.ExecutedAt = &[]time.Time{time.Now()}[0]
	proposal.UpdatedAt = time.Now()

	if err := s.governanceRepo.UpdateProposal(ctx, proposal); err != nil {
		return nil, fmt.Errorf("failed to update proposal: %w", err)
	}

	return &ProposalExecutionResult{
		ProposalID:      proposalID,
		Success:         true,
		ActionsExecuted: proposal.Actions,
		TxHashes:        txHashes,
		ExecutedAt:      time.Now(),
	}, nil
}

// Voting Management

func (s *Service) CastVote(ctx context.Context, request VoteRequest) (*entities.GovernanceVote, error) {
	// Validate vote eligibility
	eligibility, err := s.votingService.ValidateVoteEligibility(ctx, request.ProposalID, request.VoterID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate vote eligibility: %w", err)
	}

	if !eligibility.IsEligible {
		return nil, fmt.Errorf("voter not eligible: %s", eligibility.Reason)
	}

	if eligibility.AlreadyVoted {
		return nil, fmt.Errorf("voter has already voted on this proposal")
	}

	// Create vote
	vote := &entities.GovernanceVote{
		ID:           uuid.New(),
		ProposalID:   request.ProposalID,
		VoterID:      request.VoterID,
		VoterAddress: request.VoterAddress,
		Choice:       request.Choice,
		VotingPower:  eligibility.VotingPower,
		Weight:       1.0, // Default weight
		Rationale:    request.Rationale,
		Signature:    request.Signature,
		VotedAt:      time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Submit vote on-chain if signature provided
	if request.Signature != "" {
		txHash, err := s.blockchainService.SubmitVoteOnChain(ctx, vote)
		if err != nil {
			return nil, fmt.Errorf("failed to submit vote on-chain: %w", err)
		}
		vote.OnChainTxHash = txHash
	}

	// Store vote in database
	if err := s.governanceRepo.CreateVote(ctx, vote); err != nil {
		return nil, fmt.Errorf("failed to create vote: %w", err)
	}

	// Create vote cast event
	event := events.NewVoteCastEvent(vote)
	if err := s.eventRepo.Save(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create vote cast event: %v\n", err)
	}

	return vote, nil
}

func (s *Service) GetVote(ctx context.Context, proposalID, voterID uuid.UUID) (*entities.GovernanceVote, error) {
	return s.governanceRepo.GetVoteForProposalByVoter(ctx, proposalID, voterID)
}

func (s *Service) GetProposalVotes(ctx context.Context, proposalID uuid.UUID) ([]*entities.GovernanceVote, error) {
	return s.governanceRepo.GetVotesByProposal(ctx, proposalID)
}

func (s *Service) GetVoteResults(ctx context.Context, proposalID uuid.UUID) (*entities.ProposalVoteResult, error) {
	return s.governanceRepo.GetVoteResults(ctx, proposalID)
}

func (s *Service) ChangeVote(ctx context.Context, voteID uuid.UUID, newChoice entities.VoteChoice, rationale string) error {
	vote, err := s.governanceRepo.GetVoteByID(ctx, voteID)
	if err != nil {
		return fmt.Errorf("failed to get vote: %w", err)
	}

	proposal, err := s.governanceRepo.GetProposalByID(ctx, vote.ProposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal: %w", err)
	}

	if proposal.Status != entities.ProposalStatusActive {
		return fmt.Errorf("cannot change vote on proposal with status: %s", proposal.Status)
	}

	if time.Now().After(proposal.VotingEndTime) {
		return fmt.Errorf("voting period has ended")
	}

	vote.Choice = newChoice
	vote.Rationale = rationale
	vote.UpdatedAt = time.Now()

	return s.governanceRepo.UpdateVote(ctx, vote)
}

// Member Management

func (s *Service) RegisterMember(ctx context.Context, request MemberRegistrationRequest) (*entities.DAOMember, error) {
	// Validate registration request
	if err := s.membershipService.ValidateMemberRegistration(ctx, request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if member already exists
	existing, err := s.governanceRepo.GetMemberByAddress(ctx, request.Address)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("member with address %s already exists", request.Address)
	}

	// Calculate initial contribution score
	contributionScore, err := s.membershipService.CalculateContributionScore(ctx, uuid.Nil) // New member
	if err != nil {
		contributionScore = 0.0 // Default for new members
	}

	member := &entities.DAOMember{
		ID:                uuid.New(),
		Address:           request.Address,
		ENSName:           request.ENSName,
		Handle:            request.Handle,
		Role:              request.Role,
		Status:            entities.MemberStatusActive,
		TokenBalance:      request.TokenBalance,
		VotingPower:       request.TokenBalance, // Initially same as token balance
		ContributionScore: contributionScore,
		JoinedAt:          time.Now(),
		LastActivity:      time.Now(),
		VestingSchedule:   request.VestingSchedule,
		Metadata:          make(map[string]interface{}),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Store contribution proof in metadata
	if request.ContributionProof != nil {
		member.Metadata["contribution_proof"] = request.ContributionProof
	}

	// Create member in database
	if err := s.governanceRepo.CreateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to create member: %w", err)
	}

	// Create member joined event
	event := events.NewMemberJoinedEvent(member)
	if err := s.eventRepo.Save(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create member joined event: %v\n", err)
	}

	return member, nil
}

func (s *Service) GetMember(ctx context.Context, memberID uuid.UUID) (*entities.DAOMember, error) {
	return s.governanceRepo.GetMemberByID(ctx, memberID)
}

func (s *Service) GetMemberByAddress(ctx context.Context, address string) (*entities.DAOMember, error) {
	return s.governanceRepo.GetMemberByAddress(ctx, address)
}

func (s *Service) UpdateMember(ctx context.Context, memberID uuid.UUID, updates MemberUpdates) error {
	member, err := s.governanceRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	// Apply updates
	if updates.Handle != nil {
		member.Handle = *updates.Handle
	}
	if updates.Role != nil {
		member.Role = *updates.Role
	}
	if updates.Status != nil {
		member.Status = *updates.Status
	}
	if updates.ContributionScore != nil {
		member.ContributionScore = *updates.ContributionScore
	}
	if updates.VestingSchedule != nil {
		member.VestingSchedule = updates.VestingSchedule
	}
	if updates.Metadata != nil {
		for key, value := range updates.Metadata {
			member.Metadata[key] = value
		}
	}

	member.UpdatedAt = time.Now()

	return s.governanceRepo.UpdateMember(ctx, member)
}

func (s *Service) ListMembers(ctx context.Context, filter repositories.MemberFilter) ([]*entities.DAOMember, error) {
	return s.governanceRepo.ListMembers(ctx, filter)
}

func (s *Service) UpdateMemberVotingPower(ctx context.Context, memberID uuid.UUID) error {
	member, err := s.governanceRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	// Get current token balance from blockchain
	balance, err := s.blockchainService.GetTokenBalance(ctx, member.Address)
	if err != nil {
		return fmt.Errorf("failed to get token balance: %w", err)
	}

	// Get delegated voting power
	delegatedPower, err := s.GetDelegatedVotingPower(ctx, memberID)
	if err != nil {
		delegatedPower = &entities.Money{Amount: 0, Currency: balance.Currency}
	}

	// Update voting power (balance + delegated)
	totalVotingPower := &entities.Money{
		Amount:   balance.Amount + delegatedPower.Amount,
		Currency: balance.Currency,
	}

	member.TokenBalance = balance
	member.VotingPower = totalVotingPower
	member.DelegatedPower = delegatedPower
	member.LastActivity = time.Now()
	member.UpdatedAt = time.Now()

	return s.governanceRepo.UpdateMember(ctx, member)
}

// Delegation Management

func (s *Service) DelegateVotes(ctx context.Context, request DelegationRequest) (*entities.VoteDelegation, error) {
	// Validate delegation
	delegator, err := s.governanceRepo.GetMemberByID(ctx, request.DelegatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delegator: %w", err)
	}

	delegate, err := s.governanceRepo.GetMemberByID(ctx, request.DelegateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delegate: %w", err)
	}

	if delegator.Status != entities.MemberStatusActive || delegate.Status != entities.MemberStatusActive {
		return nil, fmt.Errorf("both delegator and delegate must be active")
	}

	delegation := &entities.VoteDelegation{
		ID:           uuid.New(),
		DelegatorID:  request.DelegatorID,
		DelegateID:   request.DelegateID,
		ProposalType: request.ProposalType,
		VotingPower:  delegator.VotingPower,
		IsActive:     true,
		ExpiresAt:    request.ExpiresAt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.governanceRepo.CreateDelegation(ctx, delegation); err != nil {
		return nil, fmt.Errorf("failed to create delegation: %w", err)
	}

	// Update member voting powers
	if err := s.UpdateMemberVotingPower(ctx, request.DelegatorID); err != nil {
		fmt.Printf("Failed to update delegator voting power: %v\n", err)
	}
	if err := s.UpdateMemberVotingPower(ctx, request.DelegateID); err != nil {
		fmt.Printf("Failed to update delegate voting power: %v\n", err)
	}

	return delegation, nil
}

func (s *Service) RevokeDelegation(ctx context.Context, delegationID uuid.UUID) error {
	delegation, err := s.governanceRepo.GetDelegationByID(ctx, delegationID)
	if err != nil {
		return fmt.Errorf("failed to get delegation: %w", err)
	}

	if err := s.governanceRepo.RevokeDelegation(ctx, delegationID); err != nil {
		return fmt.Errorf("failed to revoke delegation: %w", err)
	}

	// Update member voting powers
	if err := s.UpdateMemberVotingPower(ctx, delegation.DelegatorID); err != nil {
		fmt.Printf("Failed to update delegator voting power: %v\n", err)
	}
	if err := s.UpdateMemberVotingPower(ctx, delegation.DelegateID); err != nil {
		fmt.Printf("Failed to update delegate voting power: %v\n", err)
	}

	return nil
}

func (s *Service) GetDelegations(ctx context.Context, memberID uuid.UUID) ([]*entities.VoteDelegation, error) {
	return s.governanceRepo.GetDelegationsByDelegator(ctx, memberID)
}

func (s *Service) GetDelegatedVotingPower(ctx context.Context, memberID uuid.UUID) (*entities.Money, error) {
	delegations, err := s.governanceRepo.GetDelegationsByDelegate(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delegations: %w", err)
	}

	totalPower := &entities.Money{Amount: 0, Currency: "TOKEN"}
	for _, delegation := range delegations {
		if delegation.IsActive && (delegation.ExpiresAt == nil || time.Now().Before(*delegation.ExpiresAt)) {
			totalPower.Amount += delegation.VotingPower.Amount
		}
	}

	return totalPower, nil
}

// Treasury Integration (simplified implementations)

func (s *Service) CreateAllocation(ctx context.Context, request AllocationRequest) (*entities.TreasuryAllocation, error) {
	// TODO: Implement treasury service integration
	return nil, fmt.Errorf("treasury allocation creation not yet implemented")
}

func (s *Service) ExecuteAllocation(ctx context.Context, allocationID uuid.UUID) error {
	// TODO: Implement treasury service integration
	return fmt.Errorf("treasury allocation execution not yet implemented")
}

func (s *Service) GetAllocation(ctx context.Context, allocationID uuid.UUID) (*entities.TreasuryAllocation, error) {
	return s.governanceRepo.GetAllocationByID(ctx, allocationID)
}

func (s *Service) ListAllocations(ctx context.Context, filter repositories.AllocationFilter) ([]*entities.TreasuryAllocation, error) {
	return s.governanceRepo.ListAllocations(ctx, filter)
}

func (s *Service) ProcessInstallmentPayments(ctx context.Context) error {
	// TODO: Implement treasury service integration
	return fmt.Errorf("installment payment processing not yet implemented")
}

// Analytics and Reporting

func (s *Service) GetGovernanceMetrics(ctx context.Context, timeRange repositories.TimeRange) (*repositories.GovernanceMetrics, error) {
	return s.governanceRepo.GetGovernanceMetrics(ctx, timeRange)
}

func (s *Service) GetMemberParticipation(ctx context.Context, memberID uuid.UUID, timeRange repositories.TimeRange) (*repositories.MemberParticipationStats, error) {
	return s.governanceRepo.GetMemberParticipationStats(ctx, memberID, timeRange)
}

func (s *Service) GetVotingPowerDistribution(ctx context.Context) (*repositories.VotingPowerDistribution, error) {
	return s.governanceRepo.GetVotingPowerDistribution(ctx)
}

func (s *Service) GenerateGovernanceReport(ctx context.Context, request ReportRequest) (*GovernanceReport, error) {
	// This would be implemented based on the specific report type
	// For now, return a placeholder implementation
	return &GovernanceReport{
		ID:          uuid.New(),
		Type:        request.Type,
		TimeRange:   request.TimeRange,
		GeneratedAt: time.Now(),
		Format:      request.Format,
		Summary: GovernanceReportSummary{
			KeyInsights: []string{"Implementation pending"},
		},
	}, nil
}

// Configuration Management

func (s *Service) UpdateGovernanceConfig(ctx context.Context, config *entities.GovernanceConfig) error {
	config.UpdatedAt = time.Now()
	return s.governanceRepo.UpdateConfig(ctx, config)
}

func (s *Service) GetGovernanceConfig(ctx context.Context) (*entities.GovernanceConfig, error) {
	return s.governanceRepo.GetActiveConfig(ctx)
}