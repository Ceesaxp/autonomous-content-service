package dao_governance

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/google/uuid"
)

// MembershipServiceImpl implements the MembershipService interface
type MembershipServiceImpl struct {
	governanceRepo repositories.GovernanceRepository
	eventRepo      repositories.EventRepository
}

// NewMembershipService creates a new membership service instance
func NewMembershipService(
	governanceRepo repositories.GovernanceRepository,
	eventRepo repositories.EventRepository,
) *MembershipServiceImpl {
	return &MembershipServiceImpl{
		governanceRepo: governanceRepo,
		eventRepo:      eventRepo,
	}
}

// ValidateMemberRegistration validates a member registration request
func (m *MembershipServiceImpl) ValidateMemberRegistration(ctx context.Context, request MemberRegistrationRequest) error {
	// Validate address format (simplified Ethereum address validation)
	if !isValidEthereumAddress(request.Address) {
		return fmt.Errorf("invalid Ethereum address format")
	}

	// Validate handle if provided
	if request.Handle != "" {
		if err := m.validateHandle(request.Handle); err != nil {
			return fmt.Errorf("invalid handle: %w", err)
		}

		// Check if handle is already taken
		if err := m.checkHandleAvailability(ctx, request.Handle); err != nil {
			return fmt.Errorf("handle not available: %w", err)
		}
	}

	// Validate role
	if err := m.validateRole(request.Role); err != nil {
		return fmt.Errorf("invalid role: %w", err)
	}

	// Validate token balance
	if request.TokenBalance == nil || request.TokenBalance.Amount < 0 {
		return fmt.Errorf("invalid token balance")
	}

	// Validate vesting schedule if provided
	if request.VestingSchedule != nil {
		if err := m.validateVestingSchedule(*request.VestingSchedule); err != nil {
			return fmt.Errorf("invalid vesting schedule: %w", err)
		}
	}

	return nil
}

// CalculateContributionScore calculates the contribution score for a member
func (m *MembershipServiceImpl) CalculateContributionScore(ctx context.Context, memberID uuid.UUID) (float64, error) {
	if memberID == uuid.Nil {
		// New member gets default score
		return 0.0, nil
	}

	member, err := m.governanceRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get member: %w", err)
	}

	score := 0.0

	// Base score from proposals submitted
	proposalWeight := 10.0
	score += float64(member.ProposalsSubmitted) * proposalWeight

	// Score from voting participation
	votingWeight := 1.0
	score += float64(member.VotesParticipated) * votingWeight

	// Time-based bonus (longer membership = higher score)
	membershipDuration := time.Since(member.JoinedAt)
	durationBonusPerYear := 5.0
	yearsOfMembership := membershipDuration.Hours() / (24 * 365)
	score += yearsOfMembership * durationBonusPerYear

	// Role-based multiplier
	roleMultiplier := m.getRoleScoreMultiplier(member.Role)
	score *= roleMultiplier

	// Activity bonus (recent activity)
	if time.Since(member.LastActivity) < 30*24*time.Hour { // Active in last 30 days
		score *= 1.1
	}

	return score, nil
}

// UpdateMemberTokenBalance updates a member's token balance
func (m *MembershipServiceImpl) UpdateMemberTokenBalance(ctx context.Context, memberID uuid.UUID, balance *entities.Money) error {
	member, err := m.governanceRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	member.TokenBalance = balance
	member.UpdatedAt = time.Now()

	return m.governanceRepo.UpdateMember(ctx, member)
}

// ProcessVestingSchedule processes vesting for a member
func (m *MembershipServiceImpl) ProcessVestingSchedule(ctx context.Context, memberID uuid.UUID) error {
	member, err := m.governanceRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	if member.VestingSchedule == nil {
		return nil // No vesting schedule
	}

	vesting := member.VestingSchedule
	now := time.Now()

	// Check if cliff period has passed
	cliffEnd := vesting.VestingStart.Add(vesting.CliffDuration)
	if now.Before(cliffEnd) {
		return nil // Still in cliff period
	}

	// Calculate vested amount
	vestedAmount := m.calculateVestedAmount(vesting, now)
	
	// Update vested amount if it has increased
	if vestedAmount.Amount > vesting.VestedAmount.Amount {
		vesting.VestedAmount = vestedAmount
		member.VestingSchedule = vesting
		member.UpdatedAt = time.Now()

		if err := m.governanceRepo.UpdateMember(ctx, member); err != nil {
			return fmt.Errorf("failed to update member vesting: %w", err)
		}
	}

	return nil
}

// GetMemberHistory returns the activity history for a member
func (m *MembershipServiceImpl) GetMemberHistory(ctx context.Context, memberID uuid.UUID) (*MemberHistory, error) {
	member, err := m.governanceRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	history := &MemberHistory{
		MemberID: memberID,
	}

	// Get proposals submitted by member
	proposalFilter := repositories.ProposalFilter{
		ProposerID: &memberID,
		SortBy:     "created_at",
		SortOrder:  "desc",
	}
	proposals, err := m.governanceRepo.ListProposals(ctx, proposalFilter)
	if err == nil {
		history.ProposalsSubmitted = proposals
	}

	// Get votes cast by member
	voteFilter := repositories.VoteFilter{
		SortBy:    "voted_at",
		SortOrder: "desc",
	}
	votes, err := m.governanceRepo.GetVotesByVoter(ctx, memberID, voteFilter)
	if err == nil {
		history.VotesCast = votes
	}

	// Get delegations made by member
	delegationsOut, err := m.governanceRepo.GetDelegationsByDelegator(ctx, memberID)
	if err == nil {
		history.DelegationsMade = delegationsOut
	}

	// Get delegations received by member
	delegationsIn, err := m.governanceRepo.GetDelegationsByDelegate(ctx, memberID)
	if err == nil {
		history.DelegationsReceived = delegationsIn
	}

	// Create contribution history (simplified)
	history.ContributionHistory = []ContributionRecord{
		{
			Date:        member.JoinedAt,
			Type:        "membership",
			Description: "Joined DAO",
			Impact:      1.0,
		},
	}

	// Add contribution for each proposal
	for _, proposal := range proposals {
		history.ContributionHistory = append(history.ContributionHistory, ContributionRecord{
			Date:        proposal.CreatedAt,
			Type:        "proposal",
			Description: fmt.Sprintf("Submitted proposal: %s", proposal.Title),
			Impact:      10.0,
		})
	}

	// Create role change history (simplified)
	history.RoleChanges = []RoleChangeRecord{
		{
			Date:         member.JoinedAt,
			PreviousRole: entities.MemberRoleObserver,
			NewRole:      member.Role,
			Reason:       "Initial role assignment",
			ChangedBy:    memberID, // Self-assigned initially
		},
	}

	// Create token transaction history (placeholder)
	history.TokenTransactions = []TokenTransaction{
		{
			Date:        member.JoinedAt,
			Type:        "initial",
			Amount:      member.TokenBalance,
			To:          member.Address,
			Description: "Initial token allocation",
		},
	}

	return history, nil
}

// PromoteMember promotes a member to a new role
func (m *MembershipServiceImpl) PromoteMember(ctx context.Context, memberID uuid.UUID, newRole entities.MemberRole) error {
	member, err := m.governanceRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	if err := m.validateRole(newRole); err != nil {
		return fmt.Errorf("invalid new role: %w", err)
	}

	// Check if promotion is valid (simplified logic)
	if !m.isValidPromotion(member.Role, newRole) {
		return fmt.Errorf("invalid promotion from %s to %s", member.Role, newRole)
	}

	oldRole := member.Role
	member.Role = newRole
	member.UpdatedAt = time.Now()

	if err := m.governanceRepo.UpdateMember(ctx, member); err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	// Log role change in member's metadata
	roleChangeRecord := map[string]interface{}{
		"previous_role": oldRole,
		"new_role":      newRole,
		"changed_at":    time.Now(),
		"promotion":     true,
	}

	if member.Metadata == nil {
		member.Metadata = make(map[string]interface{})
	}
	
	if roleChanges, exists := member.Metadata["role_changes"]; exists {
		if changes, ok := roleChanges.([]interface{}); ok {
			member.Metadata["role_changes"] = append(changes, roleChangeRecord)
		} else {
			member.Metadata["role_changes"] = []interface{}{roleChangeRecord}
		}
	} else {
		member.Metadata["role_changes"] = []interface{}{roleChangeRecord}
	}

	return m.governanceRepo.UpdateMember(ctx, member)
}

// Helper functions

// isValidEthereumAddress validates Ethereum address format
func isValidEthereumAddress(address string) bool {
	// Basic Ethereum address validation (0x followed by 40 hex characters)
	matched, _ := regexp.MatchString("^0x[0-9a-fA-F]{40}$", address)
	return matched
}

// validateHandle validates a member handle
func (m *MembershipServiceImpl) validateHandle(handle string) error {
	if len(handle) < 3 || len(handle) > 20 {
		return fmt.Errorf("handle must be between 3 and 20 characters")
	}

	// Allow alphanumeric and underscore only
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", handle)
	if !matched {
		return fmt.Errorf("handle can only contain letters, numbers, and underscores")
	}

	// Reserved handles
	reserved := []string{"admin", "governance", "treasury", "dao", "official"}
	for _, r := range reserved {
		if strings.EqualFold(handle, r) {
			return fmt.Errorf("handle '%s' is reserved", handle)
		}
	}

	return nil
}

// checkHandleAvailability checks if a handle is available
func (m *MembershipServiceImpl) checkHandleAvailability(ctx context.Context, handle string) error {
	filter := repositories.MemberFilter{
		SearchText: handle,
	}
	
	members, err := m.governanceRepo.ListMembers(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to check handle availability: %w", err)
	}

	for _, member := range members {
		if strings.EqualFold(member.Handle, handle) {
			return fmt.Errorf("handle '%s' is already taken", handle)
		}
	}

	return nil
}

// validateRole validates a member role
func (m *MembershipServiceImpl) validateRole(role entities.MemberRole) error {
	validRoles := []entities.MemberRole{
		entities.MemberRoleFounder,
		entities.MemberRoleCore,
		entities.MemberRoleContributor,
		entities.MemberRoleDelegee,
		entities.MemberRoleObserver,
	}

	for _, validRole := range validRoles {
		if role == validRole {
			return nil
		}
	}

	return fmt.Errorf("invalid role: %s", role)
}

// validateVestingSchedule validates a vesting schedule
func (m *MembershipServiceImpl) validateVestingSchedule(schedule entities.VestingSchedule) error {
	if schedule.TotalAmount == nil || schedule.TotalAmount.Amount <= 0 {
		return fmt.Errorf("total amount must be positive")
	}

	if schedule.VestingDuration <= 0 {
		return fmt.Errorf("vesting duration must be positive")
	}

	if schedule.CliffDuration < 0 {
		return fmt.Errorf("cliff duration cannot be negative")
	}

	if schedule.CliffDuration > schedule.VestingDuration {
		return fmt.Errorf("cliff duration cannot exceed vesting duration")
	}

	if schedule.VestingStart.After(time.Now().Add(365 * 24 * time.Hour)) {
		return fmt.Errorf("vesting start cannot be more than 1 year in the future")
	}

	return nil
}

// getRoleScoreMultiplier returns the score multiplier for a role
func (m *MembershipServiceImpl) getRoleScoreMultiplier(role entities.MemberRole) float64 {
	switch role {
	case entities.MemberRoleFounder:
		return 2.0
	case entities.MemberRoleCore:
		return 1.5
	case entities.MemberRoleContributor:
		return 1.2
	case entities.MemberRoleDelegee:
		return 1.0
	case entities.MemberRoleObserver:
		return 0.8
	default:
		return 1.0
	}
}

// calculateVestedAmount calculates the vested amount at a given time
func (m *MembershipServiceImpl) calculateVestedAmount(schedule *entities.VestingSchedule, at time.Time) *entities.Money {
	if at.Before(schedule.VestingStart) {
		return &entities.Money{Amount: 0, Currency: schedule.TotalAmount.Currency}
	}

	vestingEnd := schedule.VestingStart.Add(schedule.VestingDuration)
	if at.After(vestingEnd) {
		return schedule.TotalAmount
	}

	// Linear vesting after cliff
	cliffEnd := schedule.VestingStart.Add(schedule.CliffDuration)
	if at.Before(cliffEnd) {
		return &entities.Money{Amount: 0, Currency: schedule.TotalAmount.Currency}
	}

	elapsed := at.Sub(schedule.VestingStart)
	vestedPercentage := float64(elapsed) / float64(schedule.VestingDuration)
	vestedAmount := int64(float64(schedule.TotalAmount.Amount) * vestedPercentage)

	return &entities.Money{
		Amount:   float64(vestedAmount),
		Currency: schedule.TotalAmount.Currency,
	}
}

// isValidPromotion checks if a role promotion is valid
func (m *MembershipServiceImpl) isValidPromotion(currentRole, newRole entities.MemberRole) bool {
	// Define promotion paths
	promotionPaths := map[entities.MemberRole][]entities.MemberRole{
		entities.MemberRoleObserver:    {entities.MemberRoleDelegee, entities.MemberRoleContributor},
		entities.MemberRoleDelegee:     {entities.MemberRoleContributor},
		entities.MemberRoleContributor: {entities.MemberRoleCore},
		entities.MemberRoleCore:        {entities.MemberRoleFounder},
		entities.MemberRoleFounder:     {}, // Founders cannot be promoted further
	}

	validPromotions, exists := promotionPaths[currentRole]
	if !exists {
		return false
	}

	for _, validRole := range validPromotions {
		if newRole == validRole {
			return true
		}
	}

	return false
}

// GetMembersByContributionScore returns members sorted by contribution score
func (m *MembershipServiceImpl) GetMembersByContributionScore(ctx context.Context, limit int) ([]*entities.DAOMember, error) {
	activeStatus := entities.MemberStatusActive
	filter := repositories.MemberFilter{
		Status:    &activeStatus,
		Limit:     limit,
		SortBy:    "contribution_score",
		SortOrder: "desc",
	}

	return m.governanceRepo.ListMembers(ctx, filter)
}

// UpdateMemberActivity updates a member's last activity timestamp
func (m *MembershipServiceImpl) UpdateMemberActivity(ctx context.Context, memberID uuid.UUID) error {
	member, err := m.governanceRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	member.LastActivity = time.Now()
	member.UpdatedAt = time.Now()

	return m.governanceRepo.UpdateMember(ctx, member)
}

// GetMemberStatistics returns statistics for a member
func (m *MembershipServiceImpl) GetMemberStatistics(ctx context.Context, memberID uuid.UUID) (*MemberStatistics, error) {
	member, err := m.governanceRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	// Calculate various statistics
	membershipDuration := time.Since(member.JoinedAt)
	timeSinceLastActivity := time.Since(member.LastActivity)

	stats := &MemberStatistics{
		MemberID:               memberID,
		MembershipDuration:     membershipDuration,
		TimeSinceLastActivity:  timeSinceLastActivity,
		ProposalsSubmitted:     member.ProposalsSubmitted,
		VotesParticipated:      member.VotesParticipated,
		ContributionScore:      member.ContributionScore,
		CurrentTokenBalance:    member.TokenBalance,
		CurrentVotingPower:     member.VotingPower,
		Role:                   member.Role,
		Status:                 member.Status,
	}

	// Calculate voting participation rate
	if stats.VotesParticipated > 0 && stats.ProposalsSubmitted > 0 {
		stats.VotingParticipationRate = float64(stats.VotesParticipated) / float64(stats.ProposalsSubmitted)
	}

	return stats, nil
}

// MemberStatistics represents statistics for a DAO member
type MemberStatistics struct {
	MemberID                  uuid.UUID               `json:"member_id"`
	MembershipDuration        time.Duration           `json:"membership_duration"`
	TimeSinceLastActivity     time.Duration           `json:"time_since_last_activity"`
	ProposalsSubmitted        int                     `json:"proposals_submitted"`
	VotesParticipated         int                     `json:"votes_participated"`
	VotingParticipationRate   float64                 `json:"voting_participation_rate"`
	ContributionScore         float64                 `json:"contribution_score"`
	CurrentTokenBalance       *entities.Money         `json:"current_token_balance"`
	CurrentVotingPower        *entities.Money         `json:"current_voting_power"`
	Role                      entities.MemberRole     `json:"role"`
	Status                    entities.MemberStatus   `json:"status"`
}