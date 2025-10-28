package response

type AuthResponse struct {
	PlayerID       int    `json:"player_id"`
	PlayerGUID     int    `json:"player_guid"`
	Token          string `json:"token"`
	TokenExpiredAt int64  `json:"token_expired_at"`
}
