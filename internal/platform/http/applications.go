package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/barteq100/rccc-api/internal/applications"
)

// SavedJobsHandler exposes the user-facing save-job endpoint.
type SavedJobsHandler struct {
	service *applications.Service
}

// NewSavedJobsHandler constructs the save-job handler.
func NewSavedJobsHandler(service *applications.Service) http.Handler {
	return &SavedJobsHandler{service: service}
}

// ApplicationsHandler exposes list, apply, and status-update endpoints.
type ApplicationsHandler struct {
	service *applications.Service
}

// NewApplicationsHandler constructs the applications tracking handler.
func NewApplicationsHandler(service *applications.Service) http.Handler {
	return &ApplicationsHandler{service: service}
}

type updateApplicationStatusRequest struct {
	Status string `json:"status"`
}

type listApplicationsResponse struct {
	Items []applicationResponse `json:"items"`
}

type applicationResponse struct {
	Job             jobResponse `json:"job"`
	Status          string      `json:"status"`
	SavedAt         string      `json:"saved_at"`
	AppliedAt       *string     `json:"applied_at,omitempty"`
	StatusChangedAt string      `json:"status_changed_at"`
	CreatedAt       string      `json:"created_at"`
	UpdatedAt       string      `json:"updated_at"`
}

type applicationErrorResponse struct {
	Error   string                         `json:"error"`
	Message string                         `json:"message"`
	Details []applications.ValidationIssue `json:"details,omitempty"`
}

func (h *SavedJobsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/v1/saved-jobs/") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		writeJSON(w, http.StatusMethodNotAllowed, applicationErrorResponse{
			Error:   "method_not_allowed",
			Message: "only PUT is supported",
		})
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/v1/saved-jobs/")
	application, err := h.service.Save(r.Context(), jobID)
	if err != nil {
		writeApplicationError(w, err, "failed to save job")
		return
	}

	writeJSON(w, http.StatusOK, mapApplicationResponse(application))
}

func (h *ApplicationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/v1/applications":
		h.handleCollection(w, r)
	case strings.HasPrefix(r.URL.Path, "/v1/applications/"):
		h.handleItem(w, r, strings.TrimPrefix(r.URL.Path, "/v1/applications/"))
	default:
		http.NotFound(w, r)
	}
}

func (h *ApplicationsHandler) handleCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, applicationErrorResponse{
			Error:   "method_not_allowed",
			Message: "only GET is supported",
		})
		return
	}

	items, err := h.service.List(r.Context())
	if err != nil {
		writeApplicationError(w, err, "failed to list applications")
		return
	}

	response := listApplicationsResponse{Items: make([]applicationResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapApplicationResponse(item))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *ApplicationsHandler) handleItem(w http.ResponseWriter, r *http.Request, path string) {
	if strings.HasSuffix(path, "/apply") {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeJSON(w, http.StatusMethodNotAllowed, applicationErrorResponse{
				Error:   "method_not_allowed",
				Message: "only POST is supported",
			})
			return
		}

		jobID := strings.TrimSuffix(path, "/apply")
		application, err := h.service.MarkApplied(r.Context(), jobID)
		if err != nil {
			writeApplicationError(w, err, "failed to mark job as applied")
			return
		}

		writeJSON(w, http.StatusOK, mapApplicationResponse(application))
		return
	}

	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", http.MethodPatch)
		writeJSON(w, http.StatusMethodNotAllowed, applicationErrorResponse{
			Error:   "method_not_allowed",
			Message: "only PATCH is supported",
		})
		return
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req updateApplicationStatusRequest
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, applicationErrorResponse{
			Error:   "invalid_json",
			Message: invalidJSONMessage(err),
		})
		return
	}
	if err := ensureBodyFullyConsumed(decoder); err != nil {
		writeJSON(w, http.StatusBadRequest, applicationErrorResponse{
			Error:   "invalid_json",
			Message: invalidJSONMessage(err),
		})
		return
	}

	application, err := h.service.UpdateStatus(r.Context(), path, applications.Status(strings.TrimSpace(req.Status)))
	if err != nil {
		writeApplicationError(w, err, "failed to update application status")
		return
	}

	writeJSON(w, http.StatusOK, mapApplicationResponse(application))
}

func mapApplicationResponse(item applications.TrackedApplication) applicationResponse {
	response := applicationResponse{
		Job:             mapCanonicalJobResponse(item.Job),
		Status:          string(item.Status),
		SavedAt:         item.SavedAt.UTC().Format(timeRFC3339),
		StatusChangedAt: item.StatusChangedAt.UTC().Format(timeRFC3339),
		CreatedAt:       item.CreatedAt.UTC().Format(timeRFC3339),
		UpdatedAt:       item.UpdatedAt.UTC().Format(timeRFC3339),
	}
	if item.AppliedAt != nil {
		value := item.AppliedAt.UTC().Format(timeRFC3339)
		response.AppliedAt = &value
	}

	return response
}

func writeApplicationError(w http.ResponseWriter, err error, internalMessage string) {
	var validationErr applications.ValidationError
	if errors.As(err, &validationErr) {
		writeJSON(w, http.StatusBadRequest, applicationErrorResponse{
			Error:   "validation_failed",
			Message: validationErr.Error(),
			Details: validationErr.Issues,
		})
		return
	}
	if errors.Is(err, applications.ErrJobNotFound) {
		writeJSON(w, http.StatusNotFound, applicationErrorResponse{
			Error:   "not_found",
			Message: "job not found",
		})
		return
	}

	writeJSON(w, http.StatusInternalServerError, applicationErrorResponse{
		Error:   "internal_error",
		Message: internalMessage,
	})
}
