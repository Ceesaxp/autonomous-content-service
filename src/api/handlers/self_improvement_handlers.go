package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/Ceesaxp/autonomous-content-service/src/services/self_improvement"
)

// SelfImprovementHandler handles self-improvement API endpoints
type SelfImprovementHandler struct {
	service self_improvement.SelfImprovementService
}

// NewSelfImprovementHandler creates a new self-improvement handler
func NewSelfImprovementHandler(service self_improvement.SelfImprovementService) *SelfImprovementHandler {
	return &SelfImprovementHandler{
		service: service,
	}
}

// RegisterRoutes registers all self-improvement routes
func (h *SelfImprovementHandler) RegisterRoutes(router *mux.Router) {
	// Performance monitoring
	router.HandleFunc("/api/v1/self-improvement/metrics/collect", h.CollectMetrics).Methods("POST")
	router.HandleFunc("/api/v1/self-improvement/performance/{component}", h.AnalyzePerformance).Methods("GET")
	router.HandleFunc("/api/v1/self-improvement/anomalies", h.DetectAnomalies).Methods("GET")
	
	// Learning and knowledge management
	router.HandleFunc("/api/v1/self-improvement/learn/project/{projectId}", h.LearnFromProject).Methods("POST")
	router.HandleFunc("/api/v1/self-improvement/learn/feedback/{feedbackId}", h.LearnFromFeedback).Methods("POST")
	router.HandleFunc("/api/v1/self-improvement/learn/error", h.LearnFromError).Methods("POST")
	router.HandleFunc("/api/v1/self-improvement/synthesize", h.SynthesizeLearnings).Methods("POST")
	
	// Capability management
	router.HandleFunc("/api/v1/self-improvement/capabilities/gaps", h.IdentifyCapabilityGaps).Methods("GET")
	router.HandleFunc("/api/v1/self-improvement/capabilities/gaps/{gapId}/acquire", h.AcquireCapability).Methods("POST")
	router.HandleFunc("/api/v1/self-improvement/capabilities/gaps/prioritize", h.PrioritizeGaps).Methods("POST")
	
	// Experimentation
	router.HandleFunc("/api/v1/self-improvement/experiments", h.ProposeExperiment).Methods("POST")
	router.HandleFunc("/api/v1/self-improvement/experiments/{experimentId}/run", h.RunExperiment).Methods("POST")
	router.HandleFunc("/api/v1/self-improvement/experiments/{experimentId}/evaluate", h.EvaluateExperiment).Methods("GET")
	router.HandleFunc("/api/v1/self-improvement/experiments/{experimentId}/apply", h.ApplyExperimentResults).Methods("POST")
	
	// Optimization
	router.HandleFunc("/api/v1/self-improvement/optimize/prompts/{component}", h.OptimizePrompts).Methods("POST")
	router.HandleFunc("/api/v1/self-improvement/optimize/llm", h.SelectOptimalLLM).Methods("POST")
	router.HandleFunc("/api/v1/self-improvement/optimize/workflow/{workflow}", h.OptimizeWorkflow).Methods("POST")
	
	// ROI and prioritization
	router.HandleFunc("/api/v1/self-improvement/improvements/roi", h.CalculateImprovementROI).Methods("POST")
	router.HandleFunc("/api/v1/self-improvement/improvements/prioritize", h.PrioritizeImprovements).Methods("GET")
	
	// Competitive intelligence
	router.HandleFunc("/api/v1/self-improvement/competitors/analyze", h.AnalyzeCompetitors).Methods("GET")
	router.HandleFunc("/api/v1/self-improvement/market/gaps", h.IdentifyMarketGaps).Methods("GET")
}

// Performance monitoring endpoints

func (h *SelfImprovementHandler) CollectMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if err := h.service.CollectMetrics(ctx); err != nil {
		http.Error(w, fmt.Sprintf("Failed to collect metrics: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Metrics collected successfully"})
}

func (h *SelfImprovementHandler) AnalyzePerformance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	component := vars["component"]
	
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "24h"
	}
	
	analysis, err := h.service.AnalyzePerformance(ctx, component, period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to analyze performance: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

func (h *SelfImprovementHandler) DetectAnomalies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	anomalies, err := h.service.DetectAnomalies(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to detect anomalies: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"anomalies": anomalies,
		"count":     len(anomalies),
	})
}

// Learning and knowledge management endpoints

func (h *SelfImprovementHandler) LearnFromProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	projectID := vars["projectId"]
	
	artifacts, err := h.service.LearnFromProject(ctx, projectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to learn from project: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"artifacts": artifacts,
		"count":     len(artifacts),
	})
}

func (h *SelfImprovementHandler) LearnFromFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	feedbackID := vars["feedbackId"]
	
	artifact, err := h.service.LearnFromFeedback(ctx, feedbackID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to learn from feedback: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artifact)
}

func (h *SelfImprovementHandler) LearnFromError(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var errorData self_improvement.ErrorData
	if err := json.NewDecoder(r.Body).Decode(&errorData); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	
	artifact, err := h.service.LearnFromError(ctx, errorData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to learn from error: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artifact)
}

func (h *SelfImprovementHandler) SynthesizeLearnings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var request struct {
		Period string `json:"period"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	
	if request.Period == "" {
		request.Period = "7d"
	}
	
	artifacts, err := h.service.SynthesizeLearnings(ctx, request.Period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to synthesize learnings: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"synthesized_artifacts": artifacts,
		"count":                 len(artifacts),
	})
}

// Capability management endpoints

func (h *SelfImprovementHandler) IdentifyCapabilityGaps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	gaps, err := h.service.IdentifyCapabilityGaps(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to identify capability gaps: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"gaps":  gaps,
		"count": len(gaps),
	})
}

func (h *SelfImprovementHandler) AcquireCapability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	gapID := vars["gapId"]
	
	if err := h.service.AcquireCapability(ctx, gapID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to acquire capability: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Capability acquisition initiated",
	})
}

func (h *SelfImprovementHandler) PrioritizeGaps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var request struct {
		Gaps []interface{} `json:"gaps"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	
	// For simplicity, we'll get gaps from the service instead
	gaps, err := h.service.IdentifyCapabilityGaps(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get gaps: %v", err), http.StatusInternalServerError)
		return
	}
	
	prioritizedGaps, err := h.service.PrioritizeGaps(ctx, gaps)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to prioritize gaps: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prioritized_gaps": prioritizedGaps,
		"count":            len(prioritizedGaps),
	})
}

// Experimentation endpoints

func (h *SelfImprovementHandler) ProposeExperiment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var request struct {
		Hypothesis string `json:"hypothesis"`
		Component  string `json:"component"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	
	if request.Hypothesis == "" || request.Component == "" {
		http.Error(w, "Hypothesis and component are required", http.StatusBadRequest)
		return
	}
	
	experiment, err := h.service.ProposeExperiment(ctx, request.Hypothesis, request.Component)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to propose experiment: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(experiment)
}

func (h *SelfImprovementHandler) RunExperiment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	experimentID := vars["experimentId"]
	
	if err := h.service.RunExperiment(ctx, experimentID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to run experiment: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Experiment started successfully",
	})
}

func (h *SelfImprovementHandler) EvaluateExperiment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	experimentID := vars["experimentId"]
	
	evaluation, err := h.service.EvaluateExperiment(ctx, experimentID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to evaluate experiment: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evaluation)
}

func (h *SelfImprovementHandler) ApplyExperimentResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	experimentID := vars["experimentId"]
	
	if err := h.service.ApplyExperimentResults(ctx, experimentID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply experiment results: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Experiment results applied successfully",
	})
}

// Optimization endpoints

func (h *SelfImprovementHandler) OptimizePrompts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	component := vars["component"]
	
	optimizations, err := h.service.OptimizePrompts(ctx, component)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to optimize prompts: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"optimizations": optimizations,
		"count":         len(optimizations),
	})
}

func (h *SelfImprovementHandler) SelectOptimalLLM(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var request struct {
		Task         string             `json:"task"`
		Requirements map[string]float64 `json:"requirements"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	
	if request.Task == "" {
		http.Error(w, "Task is required", http.StatusBadRequest)
		return
	}
	
	selectedLLM, err := h.service.SelectOptimalLLM(ctx, request.Task, request.Requirements)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to select optimal LLM: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"selected_llm": selectedLLM,
		"task":         request.Task,
		"requirements": request.Requirements,
	})
}

func (h *SelfImprovementHandler) OptimizeWorkflow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	workflow := vars["workflow"]
	
	optimization, err := h.service.OptimizeWorkflow(ctx, workflow)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to optimize workflow: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(optimization)
}

// ROI and prioritization endpoints

func (h *SelfImprovementHandler) CalculateImprovementROI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var improvement self_improvement.Improvement
	if err := json.NewDecoder(r.Body).Decode(&improvement); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	
	roi, err := h.service.CalculateImprovementROI(ctx, improvement)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to calculate ROI: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"improvement": improvement,
		"roi":         roi,
	})
}

func (h *SelfImprovementHandler) PrioritizeImprovements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Optional query parameters for filtering
	limitStr := r.URL.Query().Get("limit")
	limit := 20 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	improvements, err := h.service.PrioritizeImprovements(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to prioritize improvements: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Apply limit
	if len(improvements) > limit {
		improvements = improvements[:limit]
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"improvements": improvements,
		"count":        len(improvements),
		"limit":        limit,
	})
}

// Competitive intelligence endpoints

func (h *SelfImprovementHandler) AnalyzeCompetitors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	insights, err := h.service.AnalyzeCompetitors(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to analyze competitors: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"insights": insights,
		"count":    len(insights),
	})
}

func (h *SelfImprovementHandler) IdentifyMarketGaps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	opportunities, err := h.service.IdentifyMarketGaps(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to identify market gaps: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"opportunities": opportunities,
		"count":         len(opportunities),
	})
}