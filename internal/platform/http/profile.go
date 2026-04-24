package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/barteq100/rccc-api/internal/profile"
)

// ProfileHandler exposes the user-facing single-user profile preferences endpoint.
type ProfileHandler struct {
	service *profile.Service
}

// NewProfileHandler constructs the user-facing profile handler.
func NewProfileHandler(service *profile.Service) http.Handler {
	return &ProfileHandler{service: service}
}

type profilePreferencesRequest struct {
	PreferredStack     []string `json:"preferred_stack"`
	RemoteOnly         bool     `json:"remote_only"`
	PreferredLocations []string `json:"preferred_locations"`
	TargetSeniority    string   `json:"target_seniority"`
}

type profilePreferencesResponse struct {
	PreferredStack     []string `json:"preferred_stack"`
	RemoteOnly         bool     `json:"remote_only"`
	PreferredLocations []string `json:"preferred_locations"`
	TargetSeniority    string   `json:"target_seniority"`
}

type profileErrorResponse struct {
	Error   string                    `json:"error"`
	Message string                    `json:"message"`
	Details []profile.ValidationIssue `json:"details,omitempty"`
}

func (h *ProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/profile/preferences" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getPreferences(w, r)
	case http.MethodPut:
		h.putPreferences(w, r)
	default:
		w.Header().Set("Allow", "GET, PUT")
		writeJSON(w, http.StatusMethodNotAllowed, profileErrorResponse{
			Error:   "method_not_allowed",
			Message: "only GET and PUT are supported",
		})
	}
}

func (h *ProfileHandler) getPreferences(w http.ResponseWriter, r *http.Request) {
	preferences, err := h.service.Get(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, profileErrorResponse{
			Error:   "internal_error",
			Message: "failed to load profile preferences",
		})
		return
	}

	writeJSON(w, http.StatusOK, mapProfilePreferencesResponse(preferences))
}

func (h *ProfileHandler) putPreferences(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req profilePreferencesRequest
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, profileErrorResponse{
			Error:   "invalid_json",
			Message: invalidJSONMessage(err),
		})
		return
	}

	if err := ensureBodyFullyConsumed(decoder); err != nil {
		writeJSON(w, http.StatusBadRequest, profileErrorResponse{
			Error:   "invalid_json",
			Message: invalidJSONMessage(err),
		})
		return
	}

	preferences, err := h.service.Update(r.Context(), profile.UpdateInput{
		PreferredStack:     req.PreferredStack,
		RemoteOnly:         req.RemoteOnly,
		PreferredLocations: req.PreferredLocations,
		TargetSeniority:    req.TargetSeniority,
	})
	if err != nil {
		var validationErr profile.ValidationError
		if errors.As(err, &validationErr) {
			writeJSON(w, http.StatusBadRequest, profileErrorResponse{
				Error:   "validation_failed",
				Message: validationErr.Error(),
				Details: validationErr.Issues,
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, profileErrorResponse{
			Error:   "internal_error",
			Message: "failed to update profile preferences",
		})
		return
	}

	writeJSON(w, http.StatusOK, mapProfilePreferencesResponse(preferences))
}

func mapProfilePreferencesResponse(preferences profile.Preferences) profilePreferencesResponse {
	return profilePreferencesResponse{
		PreferredStack:     append([]string{}, preferences.PreferredStack...),
		RemoteOnly:         preferences.RemoteOnly,
		PreferredLocations: append([]string{}, preferences.PreferredLocations...),
		TargetSeniority:    preferences.TargetSeniority,
	}
}
