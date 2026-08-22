package auth

type ExchangeRequest struct {
	Ticket string `json:"ticket" binding:"required" example:"pQx7...base64url..."`
}
