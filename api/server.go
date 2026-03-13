package api

import (
	"net/http"

	"github.com/apparentlyarhm/app-proxy-go/config"
	"github.com/apparentlyarhm/app-proxy-go/config/middleware"
	"github.com/apparentlyarhm/app-proxy-go/internal/github"
	"github.com/apparentlyarhm/app-proxy-go/internal/report"
	"github.com/apparentlyarhm/app-proxy-go/internal/spotify"
	"github.com/apparentlyarhm/app-proxy-go/internal/steam"
	"github.com/gorilla/mux"
)

type Server struct {
	steamClient   *steam.Client // we pass the clients, with its config and hence environment details
	githubClient  *github.Client
	spotifyClient *spotify.Client
	reportClient  *report.Client
	conf          config.Config
	// We can also embed a router here
	router *mux.Router
}

func NewServer(steamClient *steam.Client, githubClient *github.Client, spotifyClient *spotify.Client, reportClient *report.Client, cfg config.Config) *Server {
	server := &Server{
		steamClient:   steamClient,
		githubClient:  githubClient,
		spotifyClient: spotifyClient,
		reportClient:  reportClient,
		conf:          cfg,
		router:        mux.NewRouter(),
	}
	server.routes()
	return server
}

// ServeHTTP makes our `Server` itself an http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.router.HandleFunc("/", s.homepageHandler()).Methods(http.MethodGet)

	s.router.HandleFunc("/steam/summary", s.handleGetSteamData()).Methods(http.MethodGet)
	s.router.HandleFunc("/github/activity", s.handleGetGithubDAta()).Methods(http.MethodGet)
	s.router.HandleFunc("/spotify/top", s.handleGetSpotifyTopItems()).Methods(http.MethodGet)
	s.router.HandleFunc("/spotify/now", s.handleGetSpotifyNowPlaying()).Methods(http.MethodGet)
	s.router.HandleFunc("/misc/ping", s.pingHandler()).Methods(http.MethodGet)

	s.router.HandleFunc("/misc/report", s.handleSystemReportRetrieval()).Methods(http.MethodGet)

	createReportHandler := http.HandlerFunc(s.handleSystemReportPublishing())
	s.router.Handle("/misc/report", middleware.WithAPIKey(s.conf.GlobalApiKey)(createReportHandler)).Methods(http.MethodPost)

	deleteReportHandler := http.HandlerFunc(s.handleSystemReportDeletion())
	s.router.Handle("/misc/report", middleware.WithAPIKey(s.conf.GlobalApiKey)(deleteReportHandler)).Methods(http.MethodDelete)

	recordViewHandler := http.HandlerFunc(s.handleViewRecording())
	s.router.Handle("/blog/record-view", middleware.WithAPIKey(s.conf.GlobalApiKey)(recordViewHandler)).Methods(http.MethodPost)
}
