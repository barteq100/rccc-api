package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/barteq100/rccc-api/internal/jobs"
)

// JobsHandler exposes the user-facing browse and detail endpoints for jobs.
type JobsHandler struct {
	service *jobs.BrowseService
}

// NewJobsHandler constructs the user-facing jobs handler.
func NewJobsHandler(service *jobs.BrowseService) http.Handler {
	return &JobsHandler{service: service}
}

type listJobsResponse struct {
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Total    int           `json:"total"`
	Items    []jobResponse `json:"items"`
}

type jobResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Remote      bool   `json:"remote"`
	Description string `json:"description"`
	Source      string `json:"source"`
	SourceURL   string `json:"source_url"`
	PostedAt    string `json:"posted_at"`
	IngestedAt  string `json:"ingested_at"`
}

func (h *JobsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error:   "method_not_allowed",
			Message: "only GET is supported",
		})
		return
	}

	if r.URL.Path == "/v1/jobs" {
		h.listJobs(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v1/jobs/") {
		h.getJob(w, r, strings.TrimPrefix(r.URL.Path, "/v1/jobs/"))
		return
	}

	http.NotFound(w, r)
}

func (h *JobsHandler) listJobs(w http.ResponseWriter, r *http.Request) {
	query, err := parseListJobsQuery(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_query",
			Message: err.Error(),
		})
		return
	}

	result, err := h.service.List(r.Context(), query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "failed to list jobs",
		})
		return
	}

	items := make([]jobResponse, 0, len(result.Items))
	for _, job := range result.Items {
		items = append(items, mapJobResponse(job))
	}

	writeJSON(w, http.StatusOK, listJobsResponse{
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
		Items:    items,
	})
}

func (h *JobsHandler) getJob(w http.ResponseWriter, r *http.Request, id string) {
	job, found, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "failed to load job",
		})
		return
	}
	if !found {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error:   "not_found",
			Message: "job not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, mapJobResponse(job))
}

func parseListJobsQuery(r *http.Request) (jobs.ListJobsQuery, error) {
	params := r.URL.Query()
	query := jobs.ListJobsQuery{
		Keyword:       params.Get("keyword"),
		Source:        params.Get("source"),
		SeniorityHint: params.Get("seniority"),
	}

	if page := params.Get("page"); page != "" {
		value, err := strconv.Atoi(page)
		if err != nil {
			return jobs.ListJobsQuery{}, errors.New("page must be an integer")
		}
		query.Page = value
	}
	if pageSize := params.Get("page_size"); pageSize != "" {
		value, err := strconv.Atoi(pageSize)
		if err != nil {
			return jobs.ListJobsQuery{}, errors.New("page_size must be an integer")
		}
		query.PageSize = value
	}
	if remote := params.Get("remote"); remote != "" {
		value, err := strconv.ParseBool(remote)
		if err != nil {
			return jobs.ListJobsQuery{}, errors.New("remote must be a boolean")
		}
		query.Remote = &value
	}

	return query, nil
}

func mapJobResponse(job jobs.Job) jobResponse {
	return jobResponse{
		ID:          job.ID,
		Title:       job.Title,
		Company:     job.Company,
		Location:    job.Location,
		Remote:      job.Remote,
		Description: job.Description,
		Source:      job.Source,
		SourceURL:   job.SourceURL,
		PostedAt:    job.PostedAt.UTC().Format(timeRFC3339),
		IngestedAt:  job.IngestedAt.UTC().Format(timeRFC3339),
	}
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"

func decodeJSONBody[T any](data []byte, target *T) error {
	return json.Unmarshal(data, target)
}
