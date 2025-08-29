package api

import (
	"net/http"

	"github.com/apparentlyarhm/app-proxy-go/internal/github"
	"github.com/apparentlyarhm/app-proxy-go/internal/steam"
)

type Server struct {
	steamClient  *steam.Client // we pass the clients, with its config and hence environment details
	githubClient *github.Client
	// We can also embed a router here
	router *http.ServeMux
}

func NewServer(steamClient *steam.Client, githubClient *github.Client) *Server {
	server := &Server{
		steamClient:  steamClient,
		githubClient: githubClient,
		router:       http.NewServeMux(),
	}
	server.routes()
	return server
}

// ServeHTTP makes our Server itself an http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.router.HandleFunc("/steam", s.handleGetSteamData())
	s.router.HandleFunc("/github/activity", s.handleGetGithubDAta())
	s.router.HandleFunc("/ping", s.pingHandler())
	// s.router.HandleFunc("/github", s.handleGetGithubData())
}
