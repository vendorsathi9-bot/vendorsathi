package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token expired")
	ErrInvalidClaims = errors.New("invalid claims")
	ErrTokenRevoked  = errors.New("token revoked")
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	UserID      string   `json:"user_id"`
	SessionID   string   `json:"session_id"`
	TokenType   string   `json:"token_type"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type JWTManager struct {
	signingKey []byte
	issuer     string
	audience   string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTManager(signingKey, issuer, audience string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		signingKey: []byte(signingKey),
		issuer:     issuer,
		audience:   audience,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *JWTManager) GenerateTokenPair(userID, sessionID string, roles, permissions []string) (*TokenPair, error) {
	now := time.Now()
	accessExpires := now.Add(m.accessTTL)
	refreshExpires := now.Add(m.refreshTTL)

	accessClaims := Claims{
		UserID:      userID,
		SessionID:   sessionID,
		TokenType:   TokenTypeAccess,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			Subject:   userID,
			ID:        uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpires),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(m.signingKey)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := Claims{
		UserID:    userID,
		SessionID: sessionID,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			Subject:   userID,
			ID:        uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpires),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(m.signingKey)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(m.accessTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, true)
}

func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, false)
}

func (m *JWTManager) validateToken(tokenString string, isAccess bool) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.signingKey, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithAudience(m.audience))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.UserID == "" || claims.SessionID == "" {
		return nil, ErrInvalidClaims
	}

	expectedType := TokenTypeAccess
	if !isAccess {
		expectedType = TokenTypeRefresh
	}
	if claims.TokenType != expectedType {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

func HashPassword(password string, memory, iterations uint32, parallelism uint8, saltLength, keyLength uint32) (string, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)

	b := make([]byte, 0, len(salt)+len(hash))
	b = append(b, salt...)
	b = append(b, hash...)

	return base64.RawStdEncoding.EncodeToString(b), nil
}

func (m *JWTManager) RefreshTTL() time.Duration {
	return m.refreshTTL
}

func VerifyPassword(password, encodedHash string, memory, iterations uint32, parallelism uint8, saltLength, keyLength uint32) (bool, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(encodedHash)
	if err != nil {
		return false, err
	}

	if len(decoded) < int(saltLength)+int(keyLength) {
		return false, errors.New("invalid hash format")
	}

	salt := decoded[:saltLength]
	hash := decoded[saltLength : saltLength+keyLength]

	expectedHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)

	return subtle.ConstantTimeCompare(hash, expectedHash) == 1, nil
}
