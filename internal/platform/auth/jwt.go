package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

const argon2idPrefix = "$argon2id$"

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
		// Pin the algorithm rather than accepting any HMAC variant. Accepting
		// e.g. HS384 with an HS256 key is an algorithm-confusion hazard.
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.signingKey, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.audience),
		jwt.WithExpirationRequired(),
	)

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

// PasswordParams holds the Argon2id cost parameters used for newly created
// password hashes. They are configurable so production deployments can tune
// cost without changing code.
type PasswordParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultPasswordParams returns the historical VendorSaathi defaults. Any
// password hash created with HashPassword (legacy, unversioned format) can be
// verified with these parameters.
func DefaultPasswordParams() PasswordParams {
	return PasswordParams{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
}

func (p PasswordParams) validate() error {
	if p.Memory == 0 || p.Iterations == 0 || p.Parallelism == 0 || p.SaltLength < 8 || p.KeyLength < 16 {
		return errors.New("invalid argon2id parameters")
	}
	return nil
}

// HashPasswordPHC hashes a password using Argon2id and returns a versioned PHC
// string: $argon2id$v=19$m=...,t=...,p=...$<salt>$<hash>. Storing the
// parameters alongside the hash allows future cost increases without breaking
// existing accounts.
func HashPasswordPHC(password string, p PasswordParams) (string, error) {
	if err := p.validate(); err != nil {
		return "", err
	}

	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	return fmt.Sprintf("%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		strings.TrimSuffix(argon2idPrefix, "$"),
		argon2.Version,
		p.Memory, p.Iterations, p.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPasswordWithParams verifies both PHC-formatted hashes and the legacy
// unversioned salt+hash format. The returned needsRehash flag is true when the
// stored hash uses fewer resources than the configured parameters (or the
// legacy format), so callers can transparently upgrade it on successful login.
func VerifyPasswordWithParams(password, encoded string, p PasswordParams) (valid bool, needsRehash bool, err error) {
	if err := p.validate(); err != nil {
		return false, false, err
	}

	if strings.HasPrefix(encoded, argon2idPrefix) {
		return verifyPHC(password, encoded, p)
	}

	// Legacy format: base64(salt || hash), no embedded parameters.
	ok, err := VerifyPassword(password, encoded, p.Memory, p.Iterations, p.Parallelism, p.SaltLength, p.KeyLength)
	if err != nil {
		return false, false, err
	}
	return ok, true, nil
}

func verifyPHC(password, encoded string, p PasswordParams) (bool, bool, error) {
	// $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, false, errors.New("invalid password hash format")
	}

	versionPart := strings.TrimPrefix(parts[2], "v=")
	version, err := strconv.Atoi(versionPart)
	if err != nil || version != argon2.Version {
		return false, false, errors.New("unsupported argon2 version")
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	for _, kv := range strings.Split(parts[3], ",") {
		pair := strings.SplitN(kv, "=", 2)
		if len(pair) != 2 {
			return false, false, errors.New("invalid argon2 parameters")
		}
		switch pair[0] {
		case "m":
			v, convErr := strconv.ParseUint(pair[1], 10, 32)
			if convErr != nil {
				return false, false, errors.New("invalid argon2 memory parameter")
			}
			memory = uint32(v)
		case "t":
			v, convErr := strconv.ParseUint(pair[1], 10, 32)
			if convErr != nil {
				return false, false, errors.New("invalid argon2 iterations parameter")
			}
			iterations = uint32(v)
		case "p":
			v, convErr := strconv.ParseUint(pair[1], 10, 8)
			if convErr != nil {
				return false, false, errors.New("invalid argon2 parallelism parameter")
			}
			parallelism = uint8(v)
		}
	}
	if memory == 0 || iterations == 0 || parallelism == 0 {
		return false, false, errors.New("invalid argon2 parameters")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) == 0 {
		return false, false, errors.New("invalid password salt")
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) == 0 {
		return false, false, errors.New("invalid password hash")
	}

	actual := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expected)))
	matches := subtle.ConstantTimeCompare(expected, actual) == 1

	needsRehash := memory < p.Memory || iterations < p.Iterations ||
		parallelism < p.Parallelism || uint32(len(salt)) < p.SaltLength ||
		uint32(len(expected)) < p.KeyLength

	return matches, needsRehash, nil
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
