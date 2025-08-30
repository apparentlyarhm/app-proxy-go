package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/apparentlyarhm/app-proxy-go/internal/spotify"
)

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

		// Parse limit and offset, with error handling
		limit, err := strconv.Atoi(queryParams.Get("limit"))
		if err != nil || limit <= 0 {
			limit = 10 // Default on error or invalid value
		}

		offset, err := strconv.Atoi(queryParams.Get("offset"))
		if err != nil || offset < 0 {
			offset = 0 // Default on error
		}

		// Parse boolean parameter
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
