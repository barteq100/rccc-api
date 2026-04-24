package main

import (
	"log"
	"net/http"
	"os"

	"github.com/barteq100/rccc-api/internal/applications"
	"github.com/barteq100/rccc-api/internal/jobs"
	transporthttp "github.com/barteq100/rccc-api/internal/platform/http"
	"github.com/barteq100/rccc-api/internal/profile"
	"github.com/barteq100/rccc-api/internal/scoring"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	repo := jobs.NewMemoryRepository()
	upsertService := jobs.NewUpsertService(repo, nil)
	profileService := profile.NewService(profile.NewMemoryRepository(), nil)
	browseService := jobs.NewBrowseService(repo, profileService, scoring.NewService())
	applicationsService := applications.NewService(applications.NewMemoryRepository(), repo, nil)

	mux := http.NewServeMux()
	mux.Handle("/internal/ingestion/jobs", transporthttp.NewJobsUpsertHandler(upsertService))
	mux.Handle("/v1/jobs", transporthttp.NewJobsHandler(browseService))
	mux.Handle("/v1/jobs/", transporthttp.NewJobsHandler(browseService))
	mux.Handle("/v1/saved-jobs/", transporthttp.NewSavedJobsHandler(applicationsService))
	mux.Handle("/v1/applications", transporthttp.NewApplicationsHandler(applicationsService))
	mux.Handle("/v1/applications/", transporthttp.NewApplicationsHandler(applicationsService))
	mux.Handle("/v1/profile/preferences", transporthttp.NewProfileHandler(profileService))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Printf("rccc-api listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
