package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
	"github.com/Ceesaxp/autonomous-content-service/src/domain/repositories"
	"github.com/Ceesaxp/autonomous-content-service/src/services/dao_governance"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// GovernanceHandlers contains handlers for DAO governance endpoints
type GovernanceHandlers struct {
	governanceService dao_governance.GovernanceService
	votingService     dao_governance.VotingService
	membershipService dao_governance.MembershipService
}

// NewGovernanceHandlers creates a new governance handlers instance
func NewGovernanceHandlers(
	governanceService dao_governance.GovernanceService,
	votingService dao_governance.VotingService,
	membershipService dao_governance.MembershipService,
) *GovernanceHandlers {
	return &GovernanceHandlers{
		governanceService: governanceService,
		votingService:     votingService,
		membershipService: membershipService,
	}
}

// RegisterRoutes registers all governance routes
func (h *GovernanceHandlers) RegisterRoutes(router *mux.Router) {
	// Proposal Management
	router.HandleFunc("/api/v1/governance/proposals", h.createProposal).Methods("POST")
	router.HandleFunc("/api/v1/governance/proposals", h.listProposals).Methods("GET")
	router.HandleFunc("/api/v1/governance/proposals/{id}", h.getProposal).Methods("GET")
	router.HandleFunc("/api/v1/governance/proposals/{id}", h.updateProposal).Methods("PUT")
	router.HandleFunc("/api/v1/governance/proposals/{id}/submit", h.submitProposal).Methods("POST")
	router.HandleFunc("/api/v1/governance/proposals/{id}/cancel", h.cancelProposal).Methods("POST")
	router.HandleFunc("/api/v1/governance/proposals/{id}/execute", h.executeProposal).Methods("POST")

	// Voting Management
	router.HandleFunc("/api/v1/governance/proposals/{id}/votes", h.castVote).Methods("POST")
	router.HandleFunc("/api/v1/governance/proposals/{id}/votes", h.getProposalVotes).Methods("GET")
	router.HandleFunc("/api/v1/governance/proposals/{id}/votes/results", h.getVoteResults).Methods("GET")
	router.HandleFunc("/api/v1/governance/votes/{id}", h.getVote).Methods("GET")
	router.HandleFunc("/api/v1/governance/votes/{id}", h.changeVote).Methods("PUT")
	router.HandleFunc("/api/v1/governance/votes/eligibility", h.checkVoteEligibility).Methods("POST")

	// Member Management
	router.HandleFunc("/api/v1/governance/members", h.registerMember).Methods("POST")
	router.HandleFunc("/api/v1/governance/members", h.listMembers).Methods("GET")
	router.HandleFunc("/api/v1/governance/members/{id}", h.getMember).Methods("GET")
	router.HandleFunc("/api/v1/governance/members/{id}", h.updateMember).Methods("PUT")
	router.HandleFunc("/api/v1/governance/members/address/{address}", h.getMemberByAddress).Methods("GET")
	router.HandleFunc("/api/v1/governance/members/{id}/promote", h.promoteMember).Methods("POST")
	router.HandleFunc("/api/v1/governance/members/{id}/history", h.getMemberHistory).Methods("GET")
	router.HandleFunc("/api/v1/governance/members/{id}/statistics", h.getMemberStatistics).Methods("GET")

	// Delegation Management
	router.HandleFunc("/api/v1/governance/delegations", h.delegateVotes).Methods("POST")
	router.HandleFunc("/api/v1/governance/delegations/{id}/revoke", h.revokeDelegation).Methods("POST")
	router.HandleFunc("/api/v1/governance/members/{id}/delegations", h.getMemberDelegations).Methods("GET")
	router.HandleFunc("/api/v1/governance/members/{id}/voting-power", h.getMemberVotingPower).Methods("GET")

	// Treasury Management
	router.HandleFunc("/api/v1/governance/treasury/allocations", h.createAllocation).Methods("POST")
	router.HandleFunc("/api/v1/governance/treasury/allocations", h.listAllocations).Methods("GET")
	router.HandleFunc("/api/v1/governance/treasury/allocations/{id}", h.getAllocation).Methods("GET")
	router.HandleFunc("/api/v1/governance/treasury/allocations/{id}/execute", h.executeAllocation).Methods("POST")
	router.HandleFunc("/api/v1/governance/treasury/process-installments", h.processInstallmentPayments).Methods("POST")

	// Analytics and Reporting
	router.HandleFunc("/api/v1/governance/metrics", h.getGovernanceMetrics).Methods("GET")
	router.HandleFunc("/api/v1/governance/members/{id}/participation", h.getMemberParticipation).Methods("GET")
	router.HandleFunc("/api/v1/governance/voting-power/distribution", h.getVotingPowerDistribution).Methods("GET")
	router.HandleFunc("/api/v1/governance/reports", h.generateGovernanceReport).Methods("POST")

	// Configuration
	router.HandleFunc("/api/v1/governance/config", h.getGovernanceConfig).Methods("GET")
	router.HandleFunc("/api/v1/governance/config", h.updateGovernanceConfig).Methods("PUT")

	// Utility endpoints
	router.HandleFunc("/api/v1/governance/proposals/active", h.getActiveProposals).Methods("GET")
	router.HandleFunc("/api/v1/governance/dashboard", h.getGovernanceDashboard).Methods("GET")
}

// Proposal Management Handlers

func (h *GovernanceHandlers) createProposal(w http.ResponseWriter, r *http.Request) {
	var request dao_governance.ProposalCreationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	proposal, err := h.governanceService.CreateProposal(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(proposal); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) listProposals(w http.ResponseWriter, r *http.Request) {
	filter := h.parseProposalFilter(r)

	proposals, err := h.governanceService.ListProposals(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"proposals": proposals,
		"total":     len(proposals),
		"filter":    filter,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getProposal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid proposal ID", http.StatusBadRequest)
		return
	}

	proposal, err := h.governanceService.GetProposal(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(proposal); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) updateProposal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid proposal ID", http.StatusBadRequest)
		return
	}

	var updates dao_governance.ProposalUpdates
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.governanceService.UpdateProposal(r.Context(), id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "updated"}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) submitProposal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid proposal ID", http.StatusBadRequest)
		return
	}

	if err := h.governanceService.SubmitProposal(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "submitted"}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) cancelProposal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid proposal ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.governanceService.CancelProposal(r.Context(), id, request.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "canceled"}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) executeProposal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid proposal ID", http.StatusBadRequest)
		return
	}

	result, err := h.governanceService.ExecuteProposal(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Voting Management Handlers

func (h *GovernanceHandlers) castVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proposalID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid proposal ID", http.StatusBadRequest)
		return
	}

	var request dao_governance.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	request.ProposalID = proposalID

	vote, err := h.governanceService.CastVote(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(vote); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getProposalVotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proposalID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid proposal ID", http.StatusBadRequest)
		return
	}

	votes, err := h.governanceService.GetProposalVotes(r.Context(), proposalID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"votes": votes,
		"total": len(votes),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getVoteResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proposalID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid proposal ID", http.StatusBadRequest)
		return
	}

	results, err := h.governanceService.GetVoteResults(r.Context(), proposalID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	voteID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid vote ID", http.StatusBadRequest)
		return
	}

	// This would need additional logic to get vote by ID
	// For now, return a placeholder response
	response := map[string]interface{}{
		"vote_id": voteID,
		"message": "Vote retrieval by ID not yet implemented",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) changeVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	voteID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid vote ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Choice    entities.VoteChoice `json:"choice"`
		Rationale string              `json:"rationale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.governanceService.ChangeVote(r.Context(), voteID, request.Choice, request.Rationale); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "updated"}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) checkVoteEligibility(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ProposalID uuid.UUID `json:"proposal_id"`
		VoterID    uuid.UUID `json:"voter_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	eligibility, err := h.votingService.ValidateVoteEligibility(r.Context(), request.ProposalID, request.VoterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(eligibility); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Member Management Handlers

func (h *GovernanceHandlers) registerMember(w http.ResponseWriter, r *http.Request) {
	var request dao_governance.MemberRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	member, err := h.governanceService.RegisterMember(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(member); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) listMembers(w http.ResponseWriter, r *http.Request) {
	filter := h.parseMemberFilter(r)

	members, err := h.governanceService.ListMembers(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"members": members,
		"total":   len(members),
		"filter":  filter,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	member, err := h.governanceService.GetMember(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(member); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getMemberByAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	member, err := h.governanceService.GetMemberByAddress(r.Context(), address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(member); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) updateMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	var updates dao_governance.MemberUpdates
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.governanceService.UpdateMember(r.Context(), id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "updated"}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) promoteMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	var request struct {
		NewRole entities.MemberRole `json:"new_role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.membershipService.PromoteMember(r.Context(), id, request.NewRole); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "promoted"}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getMemberHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	history, err := h.membershipService.GetMemberHistory(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getMemberStatistics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// This would call a method on membershipService that we haven't implemented yet
	// For now, return a placeholder response
	response := map[string]interface{}{
		"member_id": id,
		"message":   "Member statistics not yet implemented",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Helper functions for parsing query parameters

func (h *GovernanceHandlers) parseProposalFilter(r *http.Request) repositories.ProposalFilter {
	filter := repositories.ProposalFilter{}

	if status := r.URL.Query().Get("status"); status != "" {
		proposalStatus := entities.ProposalStatus(status)
		filter.Status = &proposalStatus
	}

	if proposalType := r.URL.Query().Get("type"); proposalType != "" {
		pType := entities.ProposalType(proposalType)
		filter.Type = &pType
	}

	if proposerID := r.URL.Query().Get("proposer_id"); proposerID != "" {
		if id, err := uuid.Parse(proposerID); err == nil {
			filter.ProposerID = &id
		}
	}

	if searchText := r.URL.Query().Get("search"); searchText != "" {
		filter.SearchText = searchText
	}

	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && limit > 0 {
		filter.Limit = limit
	} else {
		filter.Limit = 20 // Default limit
	}

	if offset, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && offset >= 0 {
		filter.Offset = offset
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	} else {
		filter.SortBy = "created_at"
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	} else {
		filter.SortOrder = "desc"
	}

	return filter
}

func (h *GovernanceHandlers) parseMemberFilter(r *http.Request) repositories.MemberFilter {
	filter := repositories.MemberFilter{}

	if role := r.URL.Query().Get("role"); role != "" {
		memberRole := entities.MemberRole(role)
		filter.Role = &memberRole
	}

	if status := r.URL.Query().Get("status"); status != "" {
		memberStatus := entities.MemberStatus(status)
		filter.Status = &memberStatus
	}

	if searchText := r.URL.Query().Get("search"); searchText != "" {
		filter.SearchText = searchText
	}

	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && limit > 0 {
		filter.Limit = limit
	} else {
		filter.Limit = 20 // Default limit
	}

	if offset, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && offset >= 0 {
		filter.Offset = offset
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	} else {
		filter.SortBy = "joined_at"
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	} else {
		filter.SortOrder = "desc"
	}

	return filter
}

// Placeholder implementations for remaining handlers

func (h *GovernanceHandlers) delegateVotes(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *GovernanceHandlers) revokeDelegation(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *GovernanceHandlers) getMemberDelegations(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *GovernanceHandlers) getMemberVotingPower(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *GovernanceHandlers) createAllocation(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *GovernanceHandlers) listAllocations(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *GovernanceHandlers) getAllocation(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *GovernanceHandlers) executeAllocation(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *GovernanceHandlers) processInstallmentPayments(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *GovernanceHandlers) getGovernanceMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse time range from query parameters
	startTime := time.Now().AddDate(0, -1, 0) // Default to last month
	endTime := time.Now()

	if start := r.URL.Query().Get("start"); start != "" {
		if parsed, err := time.Parse(time.RFC3339, start); err == nil {
			startTime = parsed
		}
	}

	if end := r.URL.Query().Get("end"); end != "" {
		if parsed, err := time.Parse(time.RFC3339, end); err == nil {
			endTime = parsed
		}
	}

	timeRange := repositories.TimeRange{
		Start: startTime,
		End:   endTime,
	}

	metrics, err := h.governanceService.GetGovernanceMetrics(r.Context(), timeRange)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getMemberParticipation(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *GovernanceHandlers) getVotingPowerDistribution(w http.ResponseWriter, r *http.Request) {
	distribution, err := h.governanceService.GetVotingPowerDistribution(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(distribution); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) generateGovernanceReport(w http.ResponseWriter, r *http.Request) {
	var request dao_governance.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	report, err := h.governanceService.GenerateGovernanceReport(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getGovernanceConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.governanceService.GetGovernanceConfig(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) updateGovernanceConfig(w http.ResponseWriter, r *http.Request) {
	var config entities.GovernanceConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.governanceService.UpdateGovernanceConfig(r.Context(), &config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "updated"}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getActiveProposals(w http.ResponseWriter, r *http.Request) {
	proposals, err := h.votingService.GetActiveProposalsForVoting(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"proposals": proposals,
		"total":     len(proposals),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *GovernanceHandlers) getGovernanceDashboard(w http.ResponseWriter, r *http.Request) {
	// This would aggregate various governance metrics for a dashboard view
	timeRange := repositories.TimeRange{
		Start: time.Now().AddDate(0, -1, 0), // Last month
		End:   time.Now(),
	}

	metrics, err := h.governanceService.GetGovernanceMetrics(r.Context(), timeRange)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	activeProposals, err := h.votingService.GetActiveProposalsForVoting(r.Context())
	if err != nil {
		activeProposals = []*entities.GovernanceProposal{} // Continue with empty list
	}

	dashboard := map[string]interface{}{
		"metrics":          metrics,
		"active_proposals": activeProposals,
		"summary": map[string]interface{}{
			"total_active_proposals": len(activeProposals),
			"last_updated":          time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dashboard); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}