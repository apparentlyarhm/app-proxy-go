package steam

// This is the combined response for the '?type=all' query
type AllData struct {
	Profile     ProfileResponse     `json:"profile"`
	RecentGames RecentGamesResponse `json:"recent"`
	OwnedGames  OwnedGamesResponse  `json:"games"`
}

// Corresponds to ISteamUser/GetPlayerSummaries
type ProfileResponse struct {
	Response struct {
		Players []struct {
			PersonaName  string `json:"personaname"`
			LastLogoff   int64  `json:"lastlogoff"`
			AvatarFull   string `json:"avatarfull"`
			AvatarMedium string `json:"avatarmedium"`
			Avatar       string `json:"avatar"`
			TimeCreated  int64  `json:"timecreated"`
			ProfileState int    `json:"profilestate"`
			PersonaState int    `json:"personastate"`
		} `json:"players"`
	} `json:"response"`
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
