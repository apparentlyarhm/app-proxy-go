package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GlobalRateLimit  string `envconfig:"GLOBAL_RATE_LIMIT" default:"25"`
	GlobalApiKey     string `envconfig:"GLOBAL_API_KEY" required:"true"` // as of now only the redis related apis need this to put stuff, not for reading..
	TailscaleAuthKey string `envconfig:"TAILSCALE_AUTH_KEY"`
	Database         DatabaseConfig
	Steam            SteamConfig
	Spotify          SpotifyConfig
	Github           GitHubConfig
	Redis            RedisConfig
}

type DatabaseConfig struct {
	Host string `envconfig:"DB_HOST" required:"true"`
	Name string `envconfig:"DB_NAME" required:"true"`
	User string `envconfig:"DB_USER" required:"true"`
	Pass string `envconfig:"DB_PASS" required:"true"`
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

type RedisConfig struct {
	Addr     string `envconfig:"REDIS_ADDR"`
	Password string `envconfig:"REDIS_PASSWORD"`
	DB       int    `envconfig:"REDIS_DB"`
}

func Load() (Config, error) {
	var cfg Config
	// The first argument is a prefix, which we'll leave empty.
	err := envconfig.Process("", &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to load config: %w", err)
	}

	// just protection against empty string or something, just in case...
	// using log here doesnt make sense since we are not doing anything
	fmt.Printf("[ENV] len STEAM_API_KEY: %v\n", len(cfg.Steam.APIKey))
	fmt.Printf("[ENV] len STEAM_ID: %v\n", len(cfg.Steam.ID))
	fmt.Printf("[ENV] len SPOTIFY_CLIENT_ID: %v\n", len(cfg.Spotify.ClientID))
	fmt.Printf("[ENV] len SPOTIFY_CLIENT_SECRET: %v\n", len(cfg.Spotify.ClientSecret))
	fmt.Printf("[ENV] len SPOTIFY_REFRESH_TOKEN: %v\n", len(cfg.Spotify.RefreshToken))
	fmt.Printf("[ENV] len SPOTIFY_PLAYLIST_ID: %v\n", len(cfg.Spotify.PlaylistID))
	fmt.Printf("[ENV] len GH_TOKEN: %v\n", len(cfg.Github.GhToken))
	fmt.Printf("[ENV] len REDIS_ADDR: %v\n", len(cfg.Redis.Addr))
	fmt.Printf("[ENV] len REDIS_PASSWORD: %v\n", len(cfg.Redis.Password))
	fmt.Printf("[ENV] len GLOBAL_API_KEY: %v\n", len(cfg.GlobalApiKey))
	fmt.Printf("[ENV] len DB_PASSWORD: %v\n", len(cfg.Database.Pass))
	fmt.Printf("[ENV] Setting Global Rate Limit: %v\n", cfg.GlobalRateLimit)
	fmt.Printf("[ENV] DB Host: %v\n", cfg.Database.Host)

	return cfg, nil
}
