package token

import (
	"app/internal/infrastructure/id"
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	refreshTokenTTL = 24 * time.Hour
)

type RefreshToken struct {
	Id        uuid.UUID
	Token     string
	UserId    uuid.UUID
	ExpiresAt time.Time
	Revoked   bool
}

func GenerateRefreshToken(userId uuid.UUID) (RefreshToken, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return RefreshToken{}, err
	}

	token := base64.URLEncoding.EncodeToString(b)
	token = strings.TrimRight(token, "=")

	return RefreshToken{
		Id:        id.Generate(),
		Token:     token,
		UserId:    userId,
		ExpiresAt: time.Now().Add(refreshTokenTTL),
		Revoked:   false,
	}, nil
}
