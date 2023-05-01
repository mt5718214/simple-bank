package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const minSecretKeySize = 32

var (
	ErrInvalidToken = errors.New("token is invalid")
)

type JWtMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (*JWtMaker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWtMaker{secretKey}, nil
}

func (maker *JWtMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	ss, err := jwtToken.SignedString([]byte(maker.secretKey))
	if err != nil {
		return "", err
	}

	return ss, nil
}

func (maker *JWtMaker) VerifyToken(token string) (*Payload, error) {
	keyfunc := func(token *jwt.Token) (interface{}, error) {
		/**
		* Comma-ok 斷言
		* 可以直接判斷是否是該型別的變數： value, ok = element.(T)
		* value 就是變數的值，ok 是一個 bool 型別，element 是 interface 變數，T 是斷言的型別。
		* 如果 element 裡面確實儲存了 T 型別的數值，那麼 ok 回傳 true，否則回傳 false。
		 */
		// 驗證 alg 是否為預期的HMAC演算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyfunc)
	if err != nil {
		return nil, err
	}

	if claim, ok := jwtToken.Claims.(*Payload); ok && jwtToken.Valid {
		return claim, nil
	}

	return nil, ErrInvalidToken
}

// Payload contains the payload data of the token
type Payload struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

// creates a new token payload with specific username and duration
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenId, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:       tokenId,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}

	return payload, nil
}
