package steam

// This is the combined response for the '?type=all' query
type AllData struct {
	Profile     *ProfileAPIResponse     `json:"profile"`
	RecentGames *RecentGamesApiResponse `json:"recent"`
	OwnedGames  *OwnedGamesApiResponse  `json:"games"`
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

type RecentGamesApiResponse struct {
	TotalCount int    `json:"totalCount"`
	Games      []Game `json:"games"`
	Message    string `json:"message"`
}

type OwnedGamesApiResponse struct {
	GameCount int    `json:"game_count"`
	Games     []Game `json:"games"`
}

// Corresponds to IPlayerService/GetRecentlyPlayedGames
type RecentGamesResponse struct {
	Response struct {
		TotalCount int    `json:"total_count"`
		Games      []Game `json:"games"`
	} `json:"response"`
}

// Corresponds to the RAW response from ISteamUser/GetPlayerSummaries
type ProfileResponse struct {
	Response struct {
		Players []Player `json:"players"` // Renamed for clarity
	} `json:"response"`
}

// Corresponds to IPlayerService/GetOwnedGames
type OwnedGamesResponse struct {
	Response struct {
		GameCount int    `json:"game_count"`
		Games     []Game `json:"games"`
	} `json:"response"`
}

// primitives

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

type Game struct {
	AppID           int    `json:"appid"`
	Name            string `json:"name"`
	Playtime2Weeks  int    `json:"playtime_2weeks"`
	PlaytimeForever int    `json:"playtime_forever"`
	ImgIcon         string `json:"img_icon_url"`
}

// StatusObject is the nested status object in our final response
// If a pointer is nil, it will be marshaled into the JSON null value.
type StatusObject struct {
	State  int     `json:"state"`
	InGame bool    `json:"inGame"`
	Game   *string `json:"game"`   // Use a pointer to allow for `null` in JSON
	GameID *string `json:"gameId"` // Use a pointer to allow for `null` in JSON
}
