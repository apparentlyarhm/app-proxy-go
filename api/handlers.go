package api

import (
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/apparentlyarhm/app-proxy-go/internal/report"
	"github.com/apparentlyarhm/app-proxy-go/internal/spotify"
	"github.com/redis/go-redis/v9"
)

//go:embed web/*
var content embed.FS

var homepageTemplate = template.Must(
	template.ParseFS(content, "web/index.html"),
)

var pingResponse = struct {
	Message     string `json:"message"`
	AgentString string `json:"agentString"`
}{
	Message:     "works!",
	AgentString: "go-1.25",
}

type homepageData struct {
	Redis     bool
	Service   string
	Revision  string
	StartTime string
}

func (s *Server) pingHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf(":: ping request ::")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(pingResponse)
	}

}

func (s *Server) homepageHandler() http.HandlerFunc {

	// this could be either a cold start or not, so at the frontend we will either see
	// a small time delta between this and current client time or a large, indicating it
	// has been receiving traffic.
	startTime := time.Now()

	return func(w http.ResponseWriter, r *http.Request) {

		data := homepageData{
			Redis:     os.Getenv("REDIS_ADDR") != "",
			Service:   os.Getenv("K_SERVICE"),
			Revision:  os.Getenv("K_REVISION"),
			StartTime: startTime.Format("2006-01-02 15:04:05"),
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		err := homepageTemplate.Execute(w, data)
		if err != nil {
			log.Default().Printf("%s", err)
			http.Error(w, "Failed to render homepage. However, the server is running.", http.StatusOK)
			return
		}
	}
}

func (s *Server) handleGetSteamData() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		t := r.URL.Query().Get("type")

		// We pass the client, not the raw config, to the business logic.
		data, err := s.steamClient.GetData(t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) handleGetGithubDAta() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		data, err := s.githubClient.GetGithubData()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)

	}
}

func (s *Server) handleSystemReportPublishing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p report.SystemInfo

		jErr := json.NewDecoder(r.Body).Decode(&p) // this will only fail if the payload is bad
		if jErr != nil {
			http.Error(w, "Incorrect Payload", http.StatusBadRequest)
			return
		}

		e := s.reportClient.PutSystemReport(r.Context(), p)
		if e != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleViewRecording() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var req report.ViewRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		err := s.reportClient.RecordBlogView(r.Context(), req)
		if err != nil {
			switch {
			case errors.Is(err, report.ErrInvalidSignature):
				http.Error(w, "unauthorized", http.StatusUnauthorized)

			case errors.Is(err, report.ErrMalformedPayload):
				http.Error(w, "bad request", http.StatusBadRequest)

			case errors.Is(err, report.ErrTooFast):
				http.Error(w, "view too fast", http.StatusTooEarly) // 425

			case errors.Is(err, report.ErrExpired):
				http.Error(w, "ticket expired", http.StatusGone) // 410

			case errors.Is(err, report.ErrDuplicate):
				http.Error(w, "already recorded", http.StatusConflict) // 409

			default:
				// Catch-all for internal DB/Redis errors
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"recorded"}`))
	}
}

func (s *Server) handleSystemReportRetrieval() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, e := s.reportClient.GetSystemReport(r.Context())
		if e != nil {
			// here any error is InternalServerError and is logged inside
			// however redis.Nil should be returned as 404 so that its easy to handle
			if e == redis.Nil {
				w.WriteHeader(http.StatusNotFound)
				return

			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				return

			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) handleSystemReportDeletion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e := s.reportClient.DeleteReport(r.Context())
		if e != nil {
			// here any error is InternalServerError and is logged inside
			// however redis.Nil should be returned as 404 so that its easy to handle
			if e == redis.Nil {
				w.WriteHeader(http.StatusNotFound)
				return

			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				log.Printf("Error deleting system report: %v", e)
				return

			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

// handleGetSpotifyTopItems parses query parameters for the top items endpoint.
func (s *Server) handleGetSpotifyTopItems() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()

		itemType := queryParams.Get("type")
		if itemType == "" {
			itemType = "tracks"
		}

		timeRange := queryParams.Get("time_range")
		if timeRange == "" {
			timeRange = "short_term" // Default to short_term
		}

		limit, err := strconv.Atoi(queryParams.Get("limit"))
		if err != nil || limit <= 0 {
			limit = 10 // Default on error or invalid value
		}

		offset, err := strconv.Atoi(queryParams.Get("offset"))
		if err != nil || offset < 0 {
			offset = 0 // Default on error
		}

		full, _ := strconv.ParseBool(queryParams.Get("full"))

		params := spotify.TopItemsParams{
			Type:      itemType,
			TimeRange: timeRange,
			Limit:     limit,
			Offset:    offset,
			Full:      full,
		}

		data, err := s.spotifyClient.GetTopItems(params)
		if err != nil {
			http.Error(w, "Failed to retrieve top items from Spotify.", http.StatusInternalServerError)
			log.Printf("Error retrieving top items: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	}
}

// handleGetSpotifyNowPlaying handles the request for the currently playing track.
func (s *Server) handleGetSpotifyNowPlaying() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		full, _ := strconv.ParseBool(r.URL.Query().Get("full"))

		data, err := s.spotifyClient.GetNowPlaying(full)
		if err != nil {
			http.Error(w, "Failed to retrieve now-playing data from Spotify.", http.StatusInternalServerError)
			log.Printf("Error retrieving now-playing data: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	}
}
