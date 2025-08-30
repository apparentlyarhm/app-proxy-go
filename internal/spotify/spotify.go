// I am also considering a refactor similar to how ive organized the steam.go file. This would be applicable here, but
// for now its more like a port of the express app.
package spotify

import (
	b6 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/apparentlyarhm/app-proxy-go/config"
)

type Client struct {
	config config.SpotifyConfig
	tm     TokenManager
}

func NewClient(cfg config.SpotifyConfig) *Client {

	return &Client{
		config: cfg,
		tm:     TokenManager{},
	}
}

func (c *Client) getAccessToken() (string, error) {
	// --- Fast Path: Check for valid token with read lock ---
	// A RWMutex is slightly more performant if you have many reads and few writes,
	// but a regular Mutex is perfectly fine and simpler. We'll stick with Mutex.
	c.tm.mu.Lock()
	if c.tm.accessToken != "" && time.Now().Before(c.tm.expiresAt) {
		log.Println("[Spotify] Using cached access token")
		token := c.tm.accessToken
		c.tm.mu.Unlock() // Unlock immediately after reading
		return token, nil
	}

	// We KEEP the lock to perform the refresh.
	defer c.tm.mu.Unlock()

	log.Println("[Spotify] Token is expired or missing. Refreshing...")
	authHeader := b6.StdEncoding.EncodeToString([]byte(c.config.ClientID + ":" + c.config.ClientSecret))

	params := url.Values{}
	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", c.config.RefreshToken)

	reqBody := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create token refresh request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send token refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("spotify token refresh failed with status: %s", resp.Status)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// this is the critical section
	c.tm.accessToken = tokenResp.AccessToken

	// avoid race conditions near the expiration window.
	c.tm.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	log.Println("[Spotify] New access token obtained.")

	// The lock will be released by the 'defer' statement after this return.
	return c.tm.accessToken, nil
}

func (c *Client) GetTopItems(params TopItemsParams) (any, error) {

	accessToken, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	baseURL := fmt.Sprintf("https://%s/v1/me/top/%s", c.config.Host, params.Type)
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create top items request: %w", err)
	}

	q := req.URL.Query()
	q.Add("time_range", params.TimeRange)
	q.Add("limit", strconv.Itoa(params.Limit))
	q.Add("offset", strconv.Itoa(params.Offset))
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get top items from spotify: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spotify API returned non-OK status for top items: %s", resp.Status)
	}

	var fullResponse TopItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&fullResponse); err != nil {
		return nil, fmt.Errorf("failed to decode spotify top items response: %w", err)
	}

	if params.Full {
		return fullResponse, nil
	}

	filteredItems := make([]FilteredTrack, 0, len(fullResponse.Items))
	for _, track := range fullResponse.Items {
		var image300 Image
		if len(track.Album.Images) > 0 {
			image300 = track.Album.Images[0] // Default fallback
			for _, img := range track.Album.Images {
				if img.Height == 300 && img.Width == 300 {
					image300 = img
					break
				}
			}
		}

		filteredTrack := FilteredTrack{
			Name:    track.Name,
			Artists: track.Artists,
			Album: FilteredAlbum{
				Name:   track.Album.Name,
				Images: []Image{image300}, // Create a new slice with just the one image
			},
		}
		filteredItems = append(filteredItems, filteredTrack)
	}

	// Return the final, filtered structure.
	return FilteredTopItemsResponse{Items: filteredItems}, nil
}

func (c *Client) GetNowPlaying(full bool) (any, error) {
	log.Println("[Spotify] Fetching now playing...")

	accessToken, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create now-playing request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get now-playing from spotify: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		log.Println("[Spotify] No content playing.")
		// Return an explicit "not playing" state.
		return map[string]bool{"is_playing": false}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spotify API returned non-OK status for now-playing: %s", resp.Status)
	}

	var fullResponse NowPlayingResponse
	if err := json.NewDecoder(resp.Body).Decode(&fullResponse); err != nil {
		return nil, fmt.Errorf("failed to decode spotify now-playing response: %w", err)
	}

	if full {
		return fullResponse, nil
	}

	// Ensure there's at least one image before accessing it to prevent a panic.
	var firstImage Image
	if len(fullResponse.Item.Album.Images) > 0 {
		firstImage = fullResponse.Item.Album.Images[0]
	}

	filteredResponse := FilteredNowPlayingResponse{
		IsPlaying:  fullResponse.IsPlaying,
		ProgressMs: fullResponse.ProgressMs,
		Device:     fullResponse.Device, // Device struct is simple, reuse it
		Item: FilteredPlayingItem{
			Name:         fullResponse.Item.Name,
			DurationMs:   fullResponse.Item.DurationMs,
			ExternalURLs: fullResponse.Item.ExternalURLs,
			Artists:      fullResponse.Item.Artists, // Artists are already what we want
			Album: FilteredAlbum{
				Name:   fullResponse.Item.Album.Name,
				Images: []Image{firstImage}, // Create a new slice with just the first image
			},
		},
	}

	log.Println("[Spotify] Now playing fetched successfully")
	return filteredResponse, nil
}
