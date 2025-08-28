package steam

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// before we write the actual functions, some config is required..
type SteamEnvironment struct {
	host    string
	api_key string
	id      string
}

var currentSteamEnviroment SteamEnvironment // strings get init as ""

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	currentSteamEnviroment.host = "https://api.steampowered.com"
	currentSteamEnviroment.api_key = os.Getenv("STEAM_API_KEY")
	currentSteamEnviroment.id = os.Getenv("STEAM_ID")
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
func Handler(w http.ResponseWriter, r *http.Request) {
	t := r.URL.Query().Get("type")

	data, err := getData(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// might change to something else from any...
func getProfile() (any, error) {
	// TODO: call Steam API
	return map[string]string{"stub": "profile"}, nil
}

func getRecentGames() (any, error) {
	return map[string]string{"stub": "recent games"}, nil
}

func getOwnedGames() (any, error) {
	return map[string]string{"stub": "owned games"}, nil
}

func getAll() (any, error) {
	// Could call all 3 above and merge results
	return map[string]string{"stub": "all"}, nil
}
