package spotify

import (
	"sync"
	"time"
)

// god bless ai studio to help me quickly write this, otherwise its just a pain in the ass.

type TopItemsParams struct {
	Type      string
	TimeRange string
	Limit     int
	Offset    int
	Full      bool
}

// TokenManager safely handles the Spotify token and its refresh.
// This struct would be a global variable or part of a shared app context.
type TokenManager struct {
	mu          sync.Mutex // Mutex to protect access to the fields below
	accessToken string
	expiresAt   time.Time
}

// tokenResponse defines the structure of the JSON response from Spotify's token endpoint.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// TopItemsResponse is the top-level object from the Spotify API.
type TopItemsResponse struct {
	Items []Track `json:"items"`
	// TODO: define 'artists'
}

type NowPlayingResponse struct {
	IsPlaying  bool             `json:"is_playing"`
	ProgressMs int              `json:"progress_ms"`
	Device     Device           `json:"device"`
	Item       PlayingTrackItem `json:"item"`
}

type Device struct {
	Name string `json:"name"`
	// Add other fields like ID, Type, VolumePercent if needed
}

type PlayingTrackItem struct {
	Album        Album        `json:"album"`   // We can reuse the Album struct from before
	Artists      []FullArtist `json:"artists"` // We need a more detailed Artist struct here
	Name         string       `json:"name"`
	DurationMs   int          `json:"duration_ms"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type FullArtist struct {
	Name         string       `json:"name"`
	ID           string       `json:"id"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type ExternalURLs struct {
	Spotify string `json:"spotify"`
}

// Track represents a full track object from Spotify.
type Track struct {
	Name    string   `json:"name"`
	Artists []Artist `json:"artists"`
	Album   Album    `json:"album"`
}

// Artist is a simplified artist object.
type Artist struct {
	Name string `json:"name"`
}

// Album contains album information, including a list of images.
type Album struct {
	Name   string  `json:"name"`
	Images []Image `json:"images"`
}

// Image represents an image URL with dimensions.
type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// FilteredTopItemsResponse is the custom, smaller object we'll return.
type FilteredTopItemsResponse struct {
	Items []FilteredTrack `json:"items"`
}

// FilteredTrack is the simplified track object.
type FilteredTrack struct {
	Name    string        `json:"name"`
	Artists []Artist      `json:"artists"`
	Album   FilteredAlbum `json:"album"`
}

// FilteredAlbum contains only the album name and a single image.
type FilteredAlbum struct {
	Name   string  `json:"name"`
	Images []Image `json:"images"` // We'll ensure this only has one item
}

type FilteredNowPlayingResponse struct {
	IsPlaying  bool                `json:"is_playing"`
	ProgressMs int                 `json:"progress_ms"`
	Device     Device              `json:"device"`
	Item       FilteredPlayingItem `json:"item"`
}

type FilteredPlayingItem struct {
	Album        FilteredAlbum `json:"album"`
	Artists      []FullArtist  `json:"artists"`
	Name         string        `json:"name"`
	DurationMs   int           `json:"duration_ms"`
	ExternalURLs ExternalURLs  `json:"external_urls"`
}
