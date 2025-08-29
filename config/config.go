package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Steam   SteamConfig
	Spotify SpotifyConfig
	Github  GitHubConfig
}

type SteamConfig struct {
	Host   string `envconfig:"STEAM_HOST" default:"api.steampowered.com"`
	APIKey string `envconfig:"STEAM_API_KEY" required:"true"`
	ID     string `envconfig:"STEAM_ID"      required:"true"`
}

// SpotifyConfig holds all configuration for the Spotify service.
type SpotifyConfig struct {
	Host         string `envconfig:"SPOTIFY_HOST" default:"api.spotify.com"`
	ClientID     string `envconfig:"SPOTIFY_CLIENT_ID"     required:"true"`
	ClientSecret string `envconfig:"SPOTIFY_CLIENT_SECRET" required:"true"`
	RefreshToken string `envconfig:"SPOTIFY_REFRESH_TOKEN" required:"true"`
	PlaylistID   string `envconfig:"SPOTIFY_PLAYLIST_ID"   required:"true"`
}

type GitHubConfig struct {
	Host    string `envconfig:"GH_HOST" default:"api.github.com"`
	GhToken string `envconfig:"GH_TOKEN"     required:"true"`
}

func Load() (Config, error) {
	var cfg Config
	// The first argument is a prefix, which we'll leave empty.
	err := envconfig.Process("", &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to load config: %w", err)
	}

	// just protection against empty string or something, just in case...
	fmt.Printf("[ENV] len STEAM_API_KEY: %v\n", len(cfg.Steam.APIKey))
	fmt.Printf("[ENV] len STEAM_ID: %v\n", len(cfg.Steam.ID))
	fmt.Printf("[ENV] len SPOTIFY_CLIENT_ID: %v\n", len(cfg.Spotify.ClientID))
	fmt.Printf("[ENV] len SPOTIFY_CLIENT_SECRET: %v\n", len(cfg.Spotify.ClientSecret))
	fmt.Printf("[ENV] len SPOTIFY_REFRESH_TOKEN: %v\n", len(cfg.Spotify.RefreshToken))
	fmt.Printf("[ENV] len SPOTIFY_PLAYLIST_ID: %v\n", len(cfg.Spotify.PlaylistID))
	fmt.Printf("[ENV] len GH_TOKEN: %v\n", len(cfg.Github.GhToken))

	return cfg, nil
}
