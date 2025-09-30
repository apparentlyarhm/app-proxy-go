package api

import (
	"net/http"

	"github.com/apparentlyarhm/app-proxy-go/config"
	"github.com/apparentlyarhm/app-proxy-go/config/middleware"
	"github.com/apparentlyarhm/app-proxy-go/internal/github"
	"github.com/apparentlyarhm/app-proxy-go/internal/report"
	"github.com/apparentlyarhm/app-proxy-go/internal/spotify"
	"github.com/apparentlyarhm/app-proxy-go/internal/steam"
)

type Server struct {
	steamClient   *steam.Client // we pass the clients, with its config and hence environment details
	githubClient  *github.Client
	spotifyClient *spotify.Client
	reportClient  *report.Client
	conf          config.Config
	// We can also embed a router here
	router *http.ServeMux
}

func NewServer(steamClient *steam.Client, githubClient *github.Client, spotifyClient *spotify.Client, reportClient *report.Client, cfg config.Config) *Server {
	server := &Server{
		steamClient:   steamClient,
		githubClient:  githubClient,
		spotifyClient: spotifyClient,
		reportClient:  reportClient,
		conf:          cfg,
		router:        http.NewServeMux(),
	}
	server.routes()
	return server
}

// ServeHTTP makes our `Server` itself an http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) routes() {
	// TODO: improve
	s.router.HandleFunc("/steam", s.handleGetSteamData())
	s.router.HandleFunc("/github/activity", s.handleGetGithubDAta())
	s.router.HandleFunc("/top", s.handleGetSpotifyTopItems())
	s.router.HandleFunc("/now", s.handleGetSpotifyNowPlaying())
	s.router.HandleFunc("/ping", s.pingHandler())

	getReportHandler := http.HandlerFunc(s.handleSystemReportRetrieval())
	createReportHandler := http.HandlerFunc(s.handleSystemReportPublishing())
	deleteReportHandler := http.HandlerFunc(s.handleSystemReportDeletion())

	protectedCreateReportHandler := middleware.WithAPIKey(s.conf.GlobalApiKey)(createReportHandler)
	protectedDeletedReportHandler := middleware.WithAPIKey(s.conf.GlobalApiKey)(deleteReportHandler)

	s.router.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getReportHandler.ServeHTTP(w, r)

		case http.MethodPost:
			protectedCreateReportHandler.ServeHTTP(w, r)

		case http.MethodDelete:
			protectedDeletedReportHandler.ServeHTTP(w, r)

		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
}
