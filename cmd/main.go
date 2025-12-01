package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/apparentlyarhm/app-proxy-go/api"
	"github.com/apparentlyarhm/app-proxy-go/config"
	"github.com/apparentlyarhm/app-proxy-go/internal/github"
	"github.com/apparentlyarhm/app-proxy-go/internal/report"
	"github.com/apparentlyarhm/app-proxy-go/internal/spotify"
	"github.com/apparentlyarhm/app-proxy-go/internal/steam"
	"github.com/apparentlyarhm/app-proxy-go/internal/telemetry"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// TODO: streamline error responses with proper body

	godotenv.Load() // in local the env file is present, in deployment it will get values from the enviroment directly
	fmt.Println(":: attempting to start proxy service ::")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("FATAL: could not load config: %v", err)
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://arhm.dev", "http://localhost:3000", "https://nsfw.arhm.dev"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            false,
	})

	// we also define rate limiting ops here
	// notice that the limit function returns a http.Handler. which means we can wrap our main server in it ( which is wrapped in cors as well)
	rlVal, e := strconv.Atoi(cfg.GlobalRateLimit)
	if e != nil {
		rlVal = 25 // hardcode just in case parsing fails
	}

	rl := httprate.Limit(
		rlVal,
		time.Minute,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"error": "Too many requests, please try later"}`, http.StatusTooManyRequests)
		}),
	)

	fmt.Println(`
⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⠕⠕⠕⠕⢕⢕
⢕⢕⢕⢕⢕⠕⠕⢕⢕⢕⢕⢕⢕⢕⢕⢕⢕⠕⠁⣁⣠⣤⣤⣤⣶⣦⡄⢑
⢕⢕⢕⠅⢁⣴⣤⠀⣀⠁⠑⠑⠁⢁⣀⣀⣀⣀⣘⢻⣿⣿⣿⣿⣿⡟⢁⢔
⢕⢕⠕⠀⣿⡁⠄⠀⣹⣿⣿⣿⡿⢋⣥⠤⠙⣿⣿⣿⣿⣿⡿⠿⡟⠀⢔⢕
⢕⠕⠁⣴⣦⣤⣴⣾⣿⣿⣿⣿⣇⠻⣇⠐⠀⣼⣿⣿⣿⣿⣿⣄⠀⠐⢕⢕
⠅⢀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣶⣿⣿⣿⣿⣿⣿⣿⣿⣷⡄⠐⢕
⠅⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡄⠐
⢄⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡆
⢕⢔⠀⠈⠛⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⢕⢕⢄⠈⠳⣶⣶⣶⣤⣤⣤⣤⣭⡍⢭⡍⢨⣯⡛⢿⣿⣿⣿⣿⣿⣿⣿⣿
⢕⢕⢕⢕⠀⠈⠛⠿⢿⣿⣿⣿⣿⣿⣦⣤⣿⣿⣿⣦⣈⠛⢿⢿⣿⣿⣿⣿
⢕⢕⢕⠁⢠⣾⣶⣾⣭⣖⣛⣿⠿⣿⣿⣿⣿⣿⣿⣿⣿⣷⡆⢸⣿⣿⣿⡟
⢕⠅⢀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠟⠈⢿⣿⣿⡇⡇
⢕⠕⠀⠼⠟⢉⣉⡙⠻⠿⢿⣿⣿⣿⣿⣿⡿⢿⣛⣭⡴⠶⠶⠂⠀⠿⠿⠇
	`)

	// we attempt to init redis first, as absence of it should not start the app, even though currently its not really mission critical.
	rc, e := report.NewClient(cfg.Redis)
	if e != nil {
		log.Fatalf("failed to connect to redis: %v", e)
	}

	db := config.InitDb(&cfg)

	steamHttp := &http.Client{
		Timeout:   10 * time.Second,
		Transport: &telemetry.MetricTransport{DB: db, Service: "steam", Base: http.DefaultTransport},
	}

	githubHttp := &http.Client{
		Timeout:   10 * time.Second,
		Transport: &telemetry.MetricTransport{DB: db, Service: "github", Base: http.DefaultTransport},
	}

	spotifyHttp := &http.Client{
		Timeout:   10 * time.Second,
		Transport: &telemetry.MetricTransport{DB: db, Service: "spotify", Base: http.DefaultTransport},
	}

	sc := steam.NewClient(cfg.Steam, steamHttp)        // steam client
	gc := github.NewClient(cfg.Github, githubHttp)     // github client
	spc := spotify.NewClient(cfg.Spotify, spotifyHttp) // spotify client

	server := api.NewServer(sc, gc, spc, rc, cfg) // the Server struct implements the "serveHttp" function. so its a valid http handler.

	h := c.Handler(rl(server)) // wrap server in our middleware. all requests will execute stuff from there (in our case its just adding couple of headers)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", h)) // see, its accpeted
}
