package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/barteq100/rccc-api/internal/jobs"
)

// JobsUpsertHandler exposes the ingestion-facing canonical jobs upsert endpoint.
type JobsUpsertHandler struct {
	service *jobs.UpsertService
}

// NewJobsUpsertHandler constructs the ingestion-facing upsert handler.
func NewJobsUpsertHandler(service *jobs.UpsertService) http.Handler {
	return &JobsUpsertHandler{service: service}
}

type jobsUpsertRequest struct {
	Jobs []jobs.UpsertJobInput `json:"jobs"`
}

type errorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details []jobs.ValidationIssue `json:"details,omitempty"`
}

func (h *JobsUpsertHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error:   "method_not_allowed",
			Message: "only POST is supported",
		})
		return
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req jobsUpsertRequest
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_json",
			Message: invalidJSONMessage(err),
		})
		return
	}

	if err := ensureBodyFullyConsumed(decoder); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_json",
			Message: invalidJSONMessage(err),
		})
		return
	}

	result, err := h.service.Upsert(r.Context(), req.Jobs)
	if err != nil {
		var validationErr jobs.ValidationError
		if errors.As(err, &validationErr) {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error:   "validation_failed",
				Message: validationErr.Error(),
				Details: validationErr.Issues,
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "failed to upsert jobs",
		})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func ensureBodyFullyConsumed(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	return errors.New("body must contain a single JSON object")
}

func invalidJSONMessage(err error) string {
	if errors.Is(err, io.EOF) {
		return "request body must not be empty"
	}
	return err.Error()
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
