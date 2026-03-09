package api

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/apparentlyarhm/app-proxy-go/internal/report"
	"github.com/apparentlyarhm/app-proxy-go/internal/spotify"
	"github.com/redis/go-redis/v9"
)

//go:embed web/*
var content embed.FS

var pingResponse = struct {
	Message     string `json:"message"`
	AgentString string `json:"agentString"`
}{
	Message:     "works!",
	AgentString: "go-1.25",
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

	return func(w http.ResponseWriter, r *http.Request) {

		data, err := content.ReadFile("web/index.html")
		if err != nil {
			log.Printf("Error loading homepage: %v", err)
			http.Error(w, "Failed to load homepage", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
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
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	}
}
