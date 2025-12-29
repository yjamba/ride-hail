package dto

type Response struct {
	ID           string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
