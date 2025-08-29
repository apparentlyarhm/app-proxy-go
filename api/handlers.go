package api

import (
	"encoding/json"
	"net/http"
)

var pingResponse = struct {
	Message     string
	AgentString string
}{
	Message:     "works!",
	AgentString: "go",
}

func (s *Server) pingHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
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
