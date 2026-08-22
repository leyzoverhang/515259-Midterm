package auth

type ExchangeRequest struct {
	Ticket string `json:"ticket" binding:"required" example:"pQx7...base64url..."`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required" example:"eyJhbGci..."`
}
