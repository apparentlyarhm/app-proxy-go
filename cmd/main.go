package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/apparentlyarhm/app-proxy-go/api"
	"github.com/apparentlyarhm/app-proxy-go/config"
	"github.com/apparentlyarhm/app-proxy-go/internal/github"
	"github.com/apparentlyarhm/app-proxy-go/internal/spotify"
	"github.com/apparentlyarhm/app-proxy-go/internal/steam"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load() // in local the env file is present, in deployment it will get values from the enviroment.
	fmt.Println(":: attempting to start proxy service ::")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("FATAL: could not load config: %v", err)
	}

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

	sc := steam.NewClient(cfg.Steam)      // steam client
	gc := github.NewClient(cfg.Github)    // github client
	spc := spotify.NewClient(cfg.Spotify) // spotify client

	server := api.NewServer(sc, gc, spc)
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", server))
}
