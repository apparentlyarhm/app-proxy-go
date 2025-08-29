package steam

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// before we write the actual functions, some config is required..
type SteamEnvironment struct {
	host    string
	api_key string
	id      string
}

var currentSteamEnvironment SteamEnvironment // strings get init as ""

func (p *SteamEnvironment) printLengths() {
	fmt.Printf("[STEAM-INIT] id len :: %v api_key len :: %v\n", len(p.id), len(p.api_key))
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	currentSteamEnvironment.host = "https://api.steampowered.com"
	currentSteamEnvironment.api_key = os.Getenv("STEAM_API_KEY")
	currentSteamEnvironment.id = os.Getenv("STEAM_ID")

	currentSteamEnvironment.printLengths()
}

var steamInterfaces = struct {
	USER         string
	RECENT_GAMES string
	OWNED_GAMES  string
}{
	USER:         "/ISteamUser/GetPlayerSummaries/v0002/",
	RECENT_GAMES: "/IPlayerService/GetRecentlyPlayedGames/v0001/",
	OWNED_GAMES:  "/IPlayerService/GetOwnedGames/v0001/",
}

const (
	TypeProfile  = "profile"
	TypeActivity = "activity"
	TypeOwned    = "owned"
	TypeAll      = "all"
)

// central dispatcher
func getData(t string) (any, error) {
	switch t {
	case TypeProfile:
		return getProfile()

	case TypeActivity:
		return getRecentGames()

	case TypeOwned:
		return getOwnedGames()

	case TypeAll:
		return getAll()

	default:
		return nil, errors.New("invalid 'type' param")
	}
}

// this handler will parse our query param to call appropriate fns or return error.
// notice the capital H, this can be imported anywhere else.
func Handler(w http.ResponseWriter, r *http.Request) {
	t := r.URL.Query().Get("type")

	data, err := getData(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// Sends a request to the steam api by building the URL based on inputs.
func sendRequestToSteam(steamInterface string, params map[string]string, target any) error {
	p := url.Values{}

	p.Add("key", currentSteamEnvironment.api_key)
	p.Add("format", "json")

	// we dont really need strict checking here because the params usage is a constant map only in certain places.
	for k, v := range params {
		fmt.Printf("adding %v to qs\n", k)
		p.Add(k, v)
	}

	fullURL := currentSteamEnvironment.host + steamInterface + "?" + p.Encode()
	// TODO: remove fullURL from here when everythings wrapped up.
	log.Printf("Fetching from Steam API: %s with full URL %v\n", steamInterface, fullURL)

	res, err := http.Get(fullURL)
	if err != nil {
		return fmt.Errorf("failed to send request to steam: %w", err)
	}
	// defer to guarantee the body is closed
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return fmt.Errorf("steam API responded with status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal the JSON into the target struct provided by the caller
	return json.Unmarshal(body, target)
}

func getProfile() (*ProfileAPIResponse, error) {
	var steamResponse ProfileResponse
	params := map[string]string{"steamIds": currentSteamEnvironment.id} // construct params

	// since this function unmarshals the data received into a json, we pass a reference to the `target` struct..
	err := sendRequestToSteam(steamInterfaces.USER, params, &steamResponse)
	if err != nil {
		return nil, err
	}

	player := steamResponse.Response.Players[0]
	status := StatusObject{
		State:  player.PersonaState,
		InGame: player.GameExtraInfo != "",
	}

	if status.InGame {
		status.Game = &player.GameExtraInfo
		status.GameID = &player.GameID
	}

	apiResponse := &ProfileAPIResponse{
		PersonaName:  player.PersonaName,
		LastLogoff:   player.LastLogoff,
		Avatar:       player.Avatar,
		AvatarMedium: player.AvatarMedium,
		AvatarFull:   player.AvatarFull,
		ProfileURL:   player.ProfileURL,
		TimeCreated:  player.TimeCreated,
		Status:       status,
	}
	return apiResponse, nil
}

func getRecentGames() (*RecentGamesApiResponse, error) {
	var recentSteamData RecentGamesResponse
	params := map[string]string{"steamId": currentSteamEnvironment.id}

	err := sendRequestToSteam(steamInterfaces.RECENT_GAMES, params, &recentSteamData)
	if err != nil {
		return nil, err
	}

	if recentSteamData.Response.TotalCount == 0 {
		return &RecentGamesApiResponse{
			TotalCount: 0,
			Games:      []Game{},
			Message:    "No games found",
		}, nil
	}

	apiResponse := &RecentGamesApiResponse{
		TotalCount: recentSteamData.Response.TotalCount,
		Games:      recentSteamData.Response.Games,
		Message:    "ok",
	}

	return apiResponse, err
}

func getOwnedGames() (*OwnedGamesApiResponse, error) {
	var ownedSteamData OwnedGamesResponse
	params := map[string]string{"steamId": currentSteamEnvironment.id, "include_appinfo": "true", "include_played_free_games": "true"}
	err := sendRequestToSteam(steamInterfaces.OWNED_GAMES, params, &ownedSteamData)
	if err != nil {
		return nil, err
	}

	if ownedSteamData.Response.GameCount == 0 {
		return &OwnedGamesApiResponse{
			GameCount: 0,
			Games:     []Game{},
		}, nil
	}

	apiResponse := &OwnedGamesApiResponse{
		GameCount: ownedSteamData.Response.GameCount,
		Games:     ownedSteamData.Response.Games,
	}

	return apiResponse, err

}
func getAll() (*AllData, error) {

	var wg sync.WaitGroup
	var errChan = make(chan error, 3) // Buffered channel to hold potential errors

	var allData AllData

	wg.Add(3) // We have 3 concurrent tasks to wait for

	// Goroutine for Profile
	go func() {
		defer wg.Done()
		t0 := time.Now()

		profile, err := getProfile()
		if err != nil {
			errChan <- fmt.Errorf("failed to get profile: %w", err)
			return
		}

		allData.Profile = profile
		fmt.Printf("time taken for profile :: %v\n", time.Since(t0))
	}()

	// Goroutine for Recent Games
	go func() {
		defer wg.Done()
		t0 := time.Now()

		recentGames, err := getRecentGames()
		if err != nil {
			errChan <- fmt.Errorf("failed to get recent games: %w", err)
			return
		}
		allData.RecentGames = recentGames

		fmt.Printf("time taken for recent games :: %v\n", time.Since(t0))
	}()

	// Goroutine for Owned Games
	go func() {
		defer wg.Done()
		t0 := time.Now()

		ownedGames, err := getOwnedGames()
		if err != nil {
			errChan <- fmt.Errorf("failed to get owned games: %w", err)
			return
		}
		allData.OwnedGames = ownedGames

		fmt.Printf("time taken for owned games :: %v\n", time.Since(t0))
	}()

	wg.Wait()      // Wait for all 3 goroutines to finish
	close(errChan) // Close the channel

	// Check if any errors occurred
	for err := range errChan {
		if err != nil {
			return nil, err // Return the first error we find
		}
	}

	return &allData, nil
}
