package token

import "time"

type Maker interface {
	// creates a new token for specific username and duration
	CreateToken(username string, duration time.Duration) (string, error)

	// checks if the token is vaild or not
	VerifyToken(token string) (*Payload, error)
}
