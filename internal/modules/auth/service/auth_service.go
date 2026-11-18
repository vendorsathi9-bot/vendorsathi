package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/auth/model"
	customermodel "vendorsaathi/internal/modules/customers/model"
	"vendorsaathi/internal/platform/auth"
	"vendorsaathi/internal/shared"
	"vendorsaathi/internal/platform/database"
)

// UserRepository is the interface for user data operations.
type UserRepository interface {
	CreateUser(ctx context.Context, phone, email, passwordHash string) (*model.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	AssignRole(ctx context.Context, userID, roleID uuid.UUID) error
	GetRoleByCode(ctx context.Context, code string) (*model.Role, error)
	CreateRefreshSession(ctx context.Context, session *model.RefreshSession) error
	GetRefreshSession(ctx context.Context, tokenHash string) (*model.RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, tokenHash string) error
}

// CustomerRepository is the interface for customer profile data operations.
type CustomerRepository interface {
	CreateProfile(ctx context.Context, profile *customermodel.CustomerProfile) error
	GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*customermodel.CustomerProfile, error)
}

// TxManager is the transaction manager interface.
type TxManager interface {
	RunInTx(ctx context.Context, fn func(context.Context) error) error
}

// Clock is the clock interface.
type Clock interface {
	Now() time.Time
}

// AuthService handles authentication operations for the VendorSaathi API.
type AuthService struct {
	userRepo      UserRepository
	customerRepo  CustomerRepository
	txManager     TxManager
	jwtManager    *auth.JWTManager
	clock         Clock
	dbPool        *pgxpool.Pool
	password      auth.PasswordParams
}

// NewAuthService creates a new AuthService with a database-backed login attempt tracker.
func NewAuthService(
	userRepo UserRepository,
	customerRepo CustomerRepository,
	txManager TxManager,
	jwtManager *auth.JWTManager,
	clock Clock,
	dbPool *pgxpool.Pool,
	passwordParams ...auth.PasswordParams,
) *AuthService {
	params := auth.DefaultPasswordParams()
	if len(passwordParams) > 0 {
		params = passwordParams[0]
	}
	return &AuthService{
		userRepo:     userRepo,
		customerRepo: customerRepo,
		txManager:    txManager,
		jwtManager:   jwtManager,
		clock:        clock,
		dbPool:       dbPool,
		password:     params,
	}
}

// CheckLoginAttempt checks if the account is locked due to too many failed attempts.
// Uses the database for persistence across restarts and distributed deployments.
// Returns (locked, failureCount).
func (s *AuthService) CheckLoginAttempt(ctx context.Context, phone string) (locked bool, failureCount int, err error) {
	if s.dbPool == nil {
		return false, 0, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db := database.DBFromContext(ctx, s.dbPool)
	if db == nil {
		return false, 0, nil
	}

	var count int
	var lockedUntil *time.Time
	err = db.QueryRow(ctx, `
		SELECT failed_count, locked_until FROM login_attempts
		WHERE phone = $1
	`, phone).Scan(&count, &lockedUntil)
	if err != nil {
		return false, 0, nil
	}

	if lockedUntil != nil && lockedUntil.After(s.clock.Now()) {
		return true, count, nil
	}

	return false, count, nil
}

// RecordLoginSuccess clears the failure record for the given phone number on successful login.
func (s *AuthService) RecordLoginSuccess(ctx context.Context, phone string) error {
	if s.dbPool == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db := database.DBFromContext(ctx, s.dbPool)
	if db == nil {
		return nil
	}

	_, err := db.Exec(ctx, `
		UPDATE login_attempts SET succeeded_at = NOW(), failed_count = 0, locked_until = NULL WHERE phone = $1
	`, phone)
	return err
}

// recordLoginAttemptFailure records a failed login attempt in the database.
func (s *AuthService) recordLoginAttemptFailure(ctx context.Context, phone string) {
	if s.dbPool == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db := database.DBFromContext(ctx, s.dbPool)
	if db == nil {
		return
	}

	_, _ = db.Exec(ctx, `
		INSERT INTO login_attempts (id, phone, failed_count, failed_at, locked_until, succeeded_at)
		VALUES (gen_random_uuid(), $1, 1, NOW(), NULL, NULL)
		ON CONFLICT (phone) DO UPDATE
		SET failed_count = CASE 
		        WHEN login_attempts.succeeded_at IS NOT NULL OR login_attempts.failed_at < NOW() - INTERVAL '15 minutes' THEN 1 
		        ELSE login_attempts.failed_count + 1 
		    END,
		    locked_until = CASE 
		        WHEN (login_attempts.succeeded_at IS NULL AND login_attempts.failed_at >= NOW() - INTERVAL '15 minutes' AND login_attempts.failed_count + 1 >= 5)
		        THEN NOW() + INTERVAL '15 minutes'
		        ELSE NULL 
		    END,
		    failed_at = NOW(),
		    succeeded_at = NULL
	`, phone)
}

// Login validates the user's credentials and returns a token pair.
func (s *AuthService) Login(ctx context.Context, input model.LoginInput) (*model.TokenOutput, error) {
	phone := strings.TrimSpace(input.Phone)
	if phone == "" || input.Password == "" {
		return nil, shared.NewAppError("INVALID_CREDENTIALS", "invalid credentials", shared.ErrUnauthorized)
	}

	// Check if account is locked due to too many failed attempts (database-backed)
	locked, _, err := s.CheckLoginAttempt(ctx, phone)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, shared.NewAppError("ACCOUNT_LOCKED", "too many failed login attempts; please try again later", shared.ErrUnauthorized)
	}

	user, err := s.userRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			// Phone doesn't exist — record a failure and treat as invalid credentials
			s.recordLoginAttemptFailure(ctx, phone)
			return nil, shared.NewAppError("INVALID_CREDENTIALS", "invalid credentials", shared.ErrUnauthorized)
		}
		return nil, err
	}

	valid, needsRehash, err := auth.VerifyPasswordWithParams(input.Password, user.PasswordHash, s.password)
	if err != nil || !valid {
		// Record the failure
		s.recordLoginAttemptFailure(ctx, phone)
		// Do not reveal whether the phone number exists or the hash format is
		// old; both cases are treated as invalid credentials.
		return nil, shared.NewAppError("INVALID_CREDENTIALS", "invalid credentials", shared.ErrUnauthorized)
	}

	// Record success and clear failures
	s.RecordLoginSuccess(ctx, phone)

	if needsRehash {
		// Transparently upgrade legacy/unversioned hashes after a successful
		// login. A failure here must not block authentication.
		if newHash, hashErr := auth.HashPasswordPHC(input.Password, s.password); hashErr == nil {
			if updater, ok := s.userRepo.(interface{ UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error }); ok {
				_ = updater.UpdatePassword(ctx, user.ID, newHash)
			}
		}
	}

	if user.Status != model.UserStatusActive {
		return nil, shared.NewAppError("ACCOUNT_INACTIVE", "account is not active", shared.ErrUnauthorized)
	}

	roles, err := s.userRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.userRepo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return s.generateTokenOutput(ctx, user, roles, permissions)
}

// Register creates a new user account with the given phone and password.
func (s *AuthService) Register(ctx context.Context, input model.RegisterInput) (*model.TokenOutput, error) {
	phone := strings.TrimSpace(input.Phone)
	if phone == "" {
		return nil, shared.NewValidationError("phone", "phone number is required")
	}
	if len(input.Password) < 8 {
		return nil, shared.NewValidationError("password", "password must be at least 8 characters")
	}
	email := strings.TrimSpace(input.Email)

	// 1. Check whether the phone number or email already exists.
	existingPhone, err := s.userRepo.GetUserByPhone(ctx, phone)
	if err == nil && existingPhone != nil {
		return nil, shared.NewConflict("phone already registered")
	}
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, err
	}

	if email != "" {
		existingEmail, err := s.userRepo.GetUserByEmail(ctx, email)
		if err == nil && existingEmail != nil {
			return nil, shared.NewConflict("email already registered")
		}
		if err != nil && !errors.Is(err, shared.ErrNotFound) {
			return nil, err
		}
	}

	// 2. Hash the password with Argon2id using the configured cost parameters.
	passwordHash, err := auth.HashPasswordPHC(input.Password, s.password)
	if err != nil {
		return nil, err
	}

	var user *model.User
	var tokenOutput *model.TokenOutput

	// Run database operations in a single transaction
	txErr := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		// 3. Create the user with ACTIVE status
		var createErr error
		user, createErr = s.userRepo.CreateUser(txCtx, phone, email, passwordHash)
		if createErr != nil {
			return createErr
		}

		// 4. Assign the CUSTOMER role
		customerRole, roleErr := s.userRepo.GetRoleByCode(txCtx, "CUSTOMER")
		if roleErr != nil {
			return roleErr
		}
		if assignErr := s.userRepo.AssignRole(txCtx, user.ID, customerRole.ID); assignErr != nil {
			return assignErr
		}

		// 5. Create the customer profile
		profile := &customermodel.CustomerProfile{
			UserID:             user.ID,
			ContactPreferences: []byte("{}"),
			CreatedAt:          s.clock.Now(),
			UpdatedAt:          s.clock.Now(),
		}
		if profErr := s.customerRepo.CreateProfile(txCtx, profile); profErr != nil {
			return profErr
		}

		// 6. Generate authentication tokens and persist initial session
		roles := []string{"CUSTOMER"}
		permissions, permErr := s.userRepo.GetUserPermissions(txCtx, user.ID)
		if permErr != nil {
			permissions = []string{}
		}

		var tokErr error
		tokenOutput, tokErr = s.generateTokenOutput(txCtx, user, roles, permissions)
		if tokErr != nil {
			return tokErr
		}

		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return tokenOutput, nil
}

// ResetLoginAttempt resets the failure count for the given phone number.
func (s *AuthService) ResetLoginAttempt(ctx context.Context, phone string) error {
	if s.dbPool == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db := database.DBFromContext(ctx, s.dbPool)
	if db == nil {
		return nil
	}

	_, err := db.Exec(ctx, `
		UPDATE login_attempts SET succeeded_at = NOW(), failed_count = 0, locked_until = NULL WHERE phone = $1
	`, phone)
	return err
}

// Refresh validates the refresh token JWT signature & claims.
func (s *AuthService) Refresh(ctx context.Context, input model.RefreshInput) (*model.TokenOutput, error) {
	refreshToken := strings.TrimSpace(input.RefreshToken)
	if refreshToken == "" {
		return nil, shared.NewAppError("INVALID_REFRESH_TOKEN", "refresh token is required", shared.ErrUnauthorized)
	}

	// 1. Validate the refresh token JWT signature & claims
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrExpiredToken) {
			return nil, shared.NewAppError("TOKEN_EXPIRED", "refresh token has expired", shared.ErrUnauthorized)
		}
		return nil, shared.NewAppError("INVALID_REFRESH_TOKEN", "invalid refresh token", shared.ErrUnauthorized)
	}

	// 2. Validate the refresh session in the database
	tokenHash := hashToken(refreshToken)
	session, err := s.userRepo.GetRefreshSession(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, shared.NewAppError("INVALID_REFRESH_TOKEN", "invalid refresh token", shared.ErrUnauthorized)
		}
		return nil, err
	}

	if session.RevokedAt != nil {
		return nil, shared.NewAppError("TOKEN_REVOKED", "refresh token has been revoked", shared.ErrUnauthorized)
	}

	if session.ExpiresAt.Before(s.clock.Now()) {
		return nil, shared.NewAppError("TOKEN_EXPIRED", "refresh token has expired", shared.ErrUnauthorized)
	}

	// 3. Check that the associated user is active BEFORE revoking the old session
	userUUID := session.UserID
	if claims != nil {
		if parsed, parseErr := uuid.Parse(claims.UserID); parseErr == nil {
			userUUID = parsed
		}
	}

	user, err := s.userRepo.GetUserByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, shared.NewAppError("INVALID_CREDENTIALS", "user not found", shared.ErrUnauthorized)
		}
		return nil, err
	}

	if user.Status != model.UserStatusActive {
		return nil, shared.NewAppError("ACCOUNT_INACTIVE", "account is not active", shared.ErrUnauthorized)
	}

	// 4. User is confirmed active, revoke old session
	if err := s.userRepo.RevokeRefreshSession(ctx, tokenHash); err != nil {
		return nil, err
	}

	// 5. Issue new token pair
	roles, err := s.userRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.userRepo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return s.generateTokenOutput(ctx, user, roles, permissions)
}

// Logout revokes the refresh session and logs out the user.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	return s.userRepo.RevokeRefreshSession(ctx, tokenHash)
}

// GetMe returns the current authenticated user's information.
func (s *AuthService) GetMe(ctx context.Context, userID uuid.UUID) (*model.MeOutput, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles, err := s.userRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.userRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &model.MeOutput{
		User:        user.ToUserInfo(roles),
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

// generateTokenOutput generates a token pair for the given user.
func (s *AuthService) generateTokenOutput(ctx context.Context, user *model.User, roles, permissions []string) (*model.TokenOutput, error) {
	sessionID := uuid.New().String()
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID.String(), sessionID, roles, permissions)
	if err != nil {
		return nil, err
	}

	refreshTokenHash := hashToken(tokenPair.RefreshToken)
	expiresAt := s.clock.Now().Add(s.jwtManager.RefreshTTL())

	refreshSession := &model.RefreshSession{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: expiresAt,
		DeviceID:  nil,
		CreatedAt: s.clock.Now(),
	}

	if err := s.userRepo.CreateRefreshSession(ctx, refreshSession); err != nil {
		return nil, err
	}

	return &model.TokenOutput{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		User:         user.ToUserInfo(roles),
	}, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}