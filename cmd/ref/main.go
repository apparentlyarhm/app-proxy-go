package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// this executable will not be run on prod
func main() {
	godotenv.Load()

	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		panic("SPOTIFY_CLIENT_ID or SPOTIFY_CLIENT_SECRET not set")
	}

	authCode := "code"                              // Replace with the actual authorization code obtained from Spotify's authorization flow
	redirectURI := "http://localhost:8080/callback" // Replace with your actual redirect URI

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", authCode)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest(
		http.MethodPost,
		"https://accounts.spotify.com/api/token",
		io.NopCloser(strings.NewReader(data.Encode())),
	)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	credentials := base64.StdEncoding.EncodeToString(
		[]byte(clientID + ":" + clientSecret),
	)
	req.Header.Set("Authorization", "Basic "+credentials)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.Status)
		fmt.Println(string(body))
		return
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		panic(err)
	}

	fmt.Println("Access Token:")
	fmt.Println(token.AccessToken)

	fmt.Println()

	fmt.Println("Refresh Token:")
	fmt.Println(token.RefreshToken)
}
