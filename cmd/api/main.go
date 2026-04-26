package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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

	seedDemoData, err := loadSeedDemoDataFlag()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := newServer(context.Background(), serverOptions{SeedDemoData: seedDemoData})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("rccc-api listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

type serverOptions struct {
	SeedDemoData bool
}

func newServer(ctx context.Context, options serverOptions) (http.Handler, error) {
	jobRepo := jobs.NewMemoryRepository()
	profileRepo := profile.NewMemoryRepository()
	applicationsRepo := applications.NewMemoryRepository()

	if options.SeedDemoData {
		if err := seedDemoRuntime(ctx, jobRepo, profileRepo, applicationsRepo); err != nil {
			return nil, fmt.Errorf("seed demo runtime: %w", err)
		}
	}

	upsertService := jobs.NewUpsertService(jobRepo, nil)
	profileService := profile.NewService(profileRepo, nil)
	browseService := jobs.NewBrowseService(jobRepo, profileService, scoring.NewService())
	applicationsService := applications.NewService(applicationsRepo, jobRepo, nil)

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

	return mux, nil
}

func loadSeedDemoDataFlag() (bool, error) {
	raw := strings.TrimSpace(os.Getenv("RCCC_API_SEED_DEMO_DATA"))
	if raw == "" {
		return false, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("parse RCCC_API_SEED_DEMO_DATA: %w", err)
	}

	return value, nil
}
