package steam

// This is the combined response for the '?type=all' query
type AllData struct {
	Profile     ProfileResponse     `json:"profile"`
	RecentGames RecentGamesResponse `json:"recent"`
	OwnedGames  OwnedGamesResponse  `json:"games"`
}

// Corresponds to the RAW response from ISteamUser/GetPlayerSummaries
type ProfileResponse struct {
	Response struct {
		Players []Player `json:"players"` // Renamed for clarity
	} `json:"response"`
}

type ProfileAPIResponse struct {
	PersonaName  string       `json:"personaName"`
	LastLogoff   int64        `json:"lastLogoff"`
	Avatar       string       `json:"avatar"`
	AvatarMedium string       `json:"avatarMedium"`
	AvatarFull   string       `json:"avatarFull"`
	ProfileURL   string       `json:"profileUrl"`
	TimeCreated  int64        `json:"timeCreated"`
	Status       StatusObject `json:"status"`
}

// StatusObject is the nested status object in our final response
type StatusObject struct {
	State  int     `json:"state"`
	InGame bool    `json:"inGame"`
	Game   *string `json:"game"`   // Use a pointer to allow for `null` in JSON
	GameID *string `json:"gameId"` // Use a pointer to allow for `null` in JSON
}

// Corresponds to IPlayerService/GetRecentlyPlayedGames
type RecentGamesResponse struct {
	Response struct {
		TotalCount int `json:"total_count"`
		Games      []struct {
			AppID           int    `json:"appid"`
			Name            string `json:"name"`
			Playtime2Weeks  int    `json:"playtime_2weeks"`
			PlaytimeForever int    `json:"playtime_forever"`
			ImgIcon         string `json:"img_icon_url"`
		} `json:"games"`
	} `json:"response"`
}

// Player represents a single player object from the Steam API
type Player struct {
	PersonaName   string `json:"personaname"`
	LastLogoff    int64  `json:"lastlogoff"`
	ProfileURL    string `json:"profileurl"`
	Avatar        string `json:"avatar"`
	AvatarMedium  string `json:"avatarmedium"`
	AvatarFull    string `json:"avatarfull"`
	TimeCreated   int64  `json:"timecreated"`
	PersonaState  int    `json:"personastate"`
	GameExtraInfo string `json:"gameextrainfo,omitempty"` // omitempty because it's not always present
	GameID        string `json:"gameid,omitempty"`
}

// Corresponds to IPlayerService/GetOwnedGames
type OwnedGamesResponse struct {
	Response struct {
		GameCount int `json:"game_count"`
		Games     []struct {
			AppID           int `json:"appid"`
			PlaytimeForever int `json:"playtime_forever"`
		} `json:"games"`
	} `json:"response"`
}
