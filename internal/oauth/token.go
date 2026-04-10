package oauth

import (
	"time"
)

// Token represents an OAuth2 token.
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at"`
}

// SetExpiresAt calculates and sets the ExpiresAt field based on the current time and ExpiresIn.
func (t *Token) SetExpiresAt() {
	t.ExpiresAt = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second).Unix()
}

// IsExpired checks if the token is expired or about to expire (within 5 minutes of expiry).
func (t *Token) IsExpired() bool {
	return time.Now().Unix() >= (t.ExpiresAt - 300)
}

// SetExpiresIn calculates and sets the ExpiresIn field based on the ExpiresAt field.
func (t *Token) SetExpiresIn() {
	t.ExpiresIn = int(time.Until(time.Unix(t.ExpiresAt, 0)).Seconds())
}
