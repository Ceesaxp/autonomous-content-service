package api

import (
	"github.com/Ceesaxp/autonomous-content-service/src/api/handlers"
	"github.com/gorilla/mux"
)

// SetupRoutes configures all API routes for the service
func SetupRoutes(router *mux.Router, contentHandler *handlers.ContentHandler, projectHandler *handlers.ProjectHandler) {
	// Project endpoints
	router.HandleFunc("/projects", projectHandler.CreateProject).Methods("POST")
	router.HandleFunc("/projects", projectHandler.ListProjects).Methods("GET")
	router.HandleFunc("/projects/{projectId}", projectHandler.GetProject).Methods("GET")
	router.HandleFunc("/projects/{projectId}", projectHandler.UpdateProject).Methods("PUT")
	router.HandleFunc("/projects/{projectId}", projectHandler.CancelProject).Methods("DELETE")

	// Content endpoints
	router.HandleFunc("/projects/{projectId}/content", contentHandler.CreateContent).Methods("POST")
	router.HandleFunc("/content/{contentId}", contentHandler.GetContent).Methods("GET")
	router.HandleFunc("/content/{contentId}", contentHandler.UpdateContent).Methods("PUT")
	router.HandleFunc("/content/{contentId}/versions", contentHandler.GetContentVersions).Methods("GET")
	router.HandleFunc("/content/{contentId}/approve", contentHandler.ApproveContent).Methods("POST")
}
