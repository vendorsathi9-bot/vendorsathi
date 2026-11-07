package tests

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	authmodel "vendorsaathi/internal/modules/auth/model"
	authservice "vendorsaathi/internal/modules/auth/service"
	customermodel "vendorsaathi/internal/modules/customers/model"
	"vendorsaathi/internal/platform/auth"
	"vendorsaathi/internal/shared"
)

// In-memory mock repositories and test helpers

type mockUserRepo struct {
	users           map[uuid.UUID]*authmodel.User
	usersByPhone    map[string]*authmodel.User
	usersByEmail    map[string]*authmodel.User
	userRoles       map[uuid.UUID][]string
	userPermissions map[uuid.UUID][]string
	roles           map[string]*authmodel.Role
	refreshSessions map[string]*authmodel.RefreshSession
	createUserErr   error
	assignRoleErr   error
}

func newMockUserRepo() *mockUserRepo {
	customerRoleID := uuid.New()
	adminRoleID := uuid.New()
	return &mockUserRepo{
		users:           make(map[uuid.UUID]*authmodel.User),
		usersByPhone:    make(map[string]*authmodel.User),
		usersByEmail:    make(map[string]*authmodel.User),
		userRoles:       make(map[uuid.UUID][]string),
		userPermissions: make(map[uuid.UUID][]string),
		roles: map[string]*authmodel.Role{
			"CUSTOMER": {ID: customerRoleID, Code: "CUSTOMER"},
			"ADMIN":    {ID: adminRoleID, Code: "ADMIN"},
		},
		refreshSessions: make(map[string]*authmodel.RefreshSession),
	}
}

func (m *mockUserRepo) cloneState() (map[uuid.UUID]*authmodel.User, map[string]*authmodel.User, map[string]*authmodel.User, map[uuid.UUID][]string) {
	users := make(map[uuid.UUID]*authmodel.User, len(m.users))
	for k, v := range m.users {
		u := *v
		users[k] = &u
	}
	phones := make(map[string]*authmodel.User, len(m.usersByPhone))
	for k, v := range m.usersByPhone {
		phones[k] = v
	}
	emails := make(map[string]*authmodel.User, len(m.usersByEmail))
	for k, v := range m.usersByEmail {
		emails[k] = v
	}
	roles := make(map[uuid.UUID][]string, len(m.userRoles))
	for k, v := range m.userRoles {
		r := append([]string(nil), v...)
		roles[k] = r
	}
	return users, phones, emails, roles
}

func (m *mockUserRepo) restoreState(users map[uuid.UUID]*authmodel.User, phones map[string]*authmodel.User, emails map[string]*authmodel.User, roles map[uuid.UUID][]string) {
	m.users = users
	m.usersByPhone = phones
	m.usersByEmail = emails
	m.userRoles = roles
}

func (m *mockUserRepo) CreateUser(ctx context.Context, phone, email, passwordHash string) (*authmodel.User, error) {
	if m.createUserErr != nil {
		return nil, m.createUserErr
	}
	if _, exists := m.usersByPhone[phone]; exists {
		return nil, shared.NewConflict("phone already registered")
	}
	if email != "" {
		if _, exists := m.usersByEmail[email]; exists {
			return nil, shared.NewConflict("email already registered")
		}
	}
	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}
	u := &authmodel.User{
		ID:           uuid.New(),
		Phone:        phone,
		Email:        emailPtr,
		PasswordHash: passwordHash,
		Status:       authmodel.UserStatusActive,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	m.users[u.ID] = u
	m.usersByPhone[phone] = u
	if email != "" {
		m.usersByEmail[email] = u
	}
	return u, nil
}

func (m *mockUserRepo) GetUserByPhone(ctx context.Context, phone string) (*authmodel.User, error) {
	u, ok := m.usersByPhone[phone]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*authmodel.User, error) {
	u, ok := m.usersByEmail[email]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*authmodel.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return m.userRoles[userID], nil
}

func (m *mockUserRepo) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	perms := m.userPermissions[userID]
	if perms == nil {
		return []string{"product:read", "order:read", "order:write"}, nil
	}
	return perms, nil
}

func (m *mockUserRepo) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.assignRoleErr != nil {
		return m.assignRoleErr
	}
	for _, r := range m.roles {
		if r.ID == roleID {
			m.userRoles[userID] = append(m.userRoles[userID], r.Code)
			return nil
		}
	}
	return shared.ErrNotFound
}

func (m *mockUserRepo) GetRoleByCode(ctx context.Context, code string) (*authmodel.Role, error) {
	r, ok := m.roles[code]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return r, nil
}

func (m *mockUserRepo) CreateRefreshSession(ctx context.Context, session *authmodel.RefreshSession) error {
	m.refreshSessions[session.TokenHash] = session
	return nil
}

func (m *mockUserRepo) GetRefreshSession(ctx context.Context, tokenHash string) (*authmodel.RefreshSession, error) {
	s, ok := m.refreshSessions[tokenHash]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return s, nil
}

func (m *mockUserRepo) RevokeRefreshSession(ctx context.Context, tokenHash string) error {
	s, ok := m.refreshSessions[tokenHash]
	if !ok {
		return shared.ErrNotFound
	}
	now := time.Now().UTC()
	s.RevokedAt = &now
	return nil
}

type mockCustomerRepo struct {
	profiles         map[uuid.UUID]*customermodel.CustomerProfile
	createProfileErr error
}

func newMockCustomerRepo() *mockCustomerRepo {
	return &mockCustomerRepo{
		profiles: make(map[uuid.UUID]*customermodel.CustomerProfile),
	}
}

func (m *mockCustomerRepo) cloneState() map[uuid.UUID]*customermodel.CustomerProfile {
	res := make(map[uuid.UUID]*customermodel.CustomerProfile, len(m.profiles))
	for k, v := range m.profiles {
		p := *v
		res[k] = &p
	}
	return res
}

func (m *mockCustomerRepo) restoreState(state map[uuid.UUID]*customermodel.CustomerProfile) {
	m.profiles = state
}

func (m *mockCustomerRepo) CreateProfile(ctx context.Context, profile *customermodel.CustomerProfile) error {
	if m.createProfileErr != nil {
		return m.createProfileErr
	}
	m.profiles[profile.UserID] = profile
	return nil
}

func (m *mockCustomerRepo) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*customermodel.CustomerProfile, error) {
	p, ok := m.profiles[userID]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return p, nil
}

type mockTxManager struct {
	userRepo     *mockUserRepo
	customerRepo *mockCustomerRepo
	rolledBack   bool
	committed    bool
}

func (tm *mockTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	uUsers, uPhones, uEmails, uRoles := tm.userRepo.cloneState()
	cProfiles := tm.customerRepo.cloneState()

	err := fn(ctx)
	if err != nil {
		tm.rolledBack = true
		tm.userRepo.restoreState(uUsers, uPhones, uEmails, uRoles)
		tm.customerRepo.restoreState(cProfiles)
		return err
	}
	tm.committed = true
	return nil
}

type fixedClock struct {
	now time.Time
}

func (c *fixedClock) Now() time.Time {
	return c.now
}

func setupAuthTestEnv() (*authservice.AuthService, *mockUserRepo, *mockCustomerRepo, *mockTxManager, *auth.JWTManager, *fixedClock) {
	uRepo := newMockUserRepo()
	cRepo := newMockCustomerRepo()
	txMgr := &mockTxManager{userRepo: uRepo, customerRepo: cRepo}
	jwtMgr := auth.NewJWTManager("test-secret-key-at-least-32-bytes-long!", "vendorsaathi", "vendorsaathi-test", 15*time.Minute, 24*time.Hour)
	clk := &fixedClock{now: time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)}
	svc := authservice.NewAuthService(uRepo, cRepo, txMgr, jwtMgr, clk)
	return svc, uRepo, cRepo, txMgr, jwtMgr, clk
}

// 1. Successful customer registration
func TestRegister_Success(t *testing.T) {
	svc, uRepo, cRepo, txMgr, jwtMgr, _ := setupAuthTestEnv()
	ctx := context.Background()

	input := authmodel.RegisterInput{
		Phone:    "+919876543210",
		Email:    "test@example.com",
		Password: "SecurePassword123",
	}

	out, err := svc.Register(ctx, input)
	if err != nil {
		t.Fatalf("expected successful registration, got: %v", err)
	}

	if out == nil {
		t.Fatal("expected token output, got nil")
	}
	if out.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if out.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if out.TokenType != "Bearer" {
		t.Errorf("expected token type Bearer, got %s", out.TokenType)
	}
	if out.User.Phone != input.Phone {
		t.Errorf("expected phone %s, got %s", input.Phone, out.User.Phone)
	}
	if out.User.Status != authmodel.UserStatusActive {
		t.Errorf("expected status ACTIVE, got %s", out.User.Status)
	}
	if len(out.User.Roles) != 1 || out.User.Roles[0] != "CUSTOMER" {
		t.Errorf("expected role [CUSTOMER], got %v", out.User.Roles)
	}

	// Validate access token claims
	claims, err := jwtMgr.ValidateAccessToken(out.AccessToken)
	if err != nil {
		t.Fatalf("failed to validate returned access token: %v", err)
	}
	if claims.UserID != out.User.ID.String() {
		t.Errorf("expected claim user id %s, got %s", out.User.ID.String(), claims.UserID)
	}

	// Verify persistence in user repository
	savedUser, err := uRepo.GetUserByID(ctx, out.User.ID)
	if err != nil {
		t.Fatalf("user was not saved in repo: %v", err)
	}
	if savedUser.Status != authmodel.UserStatusActive {
		t.Errorf("saved user status should be ACTIVE, got %s", savedUser.Status)
	}

	// Verify customer profile created
	profile, err := cRepo.GetProfileByUserID(ctx, out.User.ID)
	if err != nil {
		t.Fatalf("customer profile was not created: %v", err)
	}
	if profile.UserID != out.User.ID {
		t.Errorf("profile user_id mismatch: expected %s, got %s", out.User.ID, profile.UserID)
	}

	// Verify transaction committed
	if !txMgr.committed {
		t.Error("expected transaction to be committed")
	}
}

// 2. Duplicate phone number
func TestRegister_DuplicatePhone(t *testing.T) {
	svc, _, _, _, _, _ := setupAuthTestEnv()
	ctx := context.Background()

	input := authmodel.RegisterInput{
		Phone:    "+919876543210",
		Email:    "user1@example.com",
		Password: "Password123",
	}

	_, err := svc.Register(ctx, input)
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	// Try to register same phone with different email
	input2 := authmodel.RegisterInput{
		Phone:    "+919876543210",
		Email:    "user2@example.com",
		Password: "Password123",
	}
	_, err2 := svc.Register(ctx, input2)
	if err2 == nil {
		t.Fatal("expected conflict error on duplicate phone, got nil")
	}
	if !shared.IsConflict(err2) {
		t.Errorf("expected conflict error, got: %v", err2)
	}
}

// 3. Duplicate email, when supplied
func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _, _, _, _, _ := setupAuthTestEnv()
	ctx := context.Background()

	input1 := authmodel.RegisterInput{
		Phone:    "+919876543210",
		Email:    "shared@example.com",
		Password: "Password123",
	}
	_, err := svc.Register(ctx, input1)
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	// Register with different phone but duplicate email
	input2 := authmodel.RegisterInput{
		Phone:    "+919876543211",
		Email:    "shared@example.com",
		Password: "Password123",
	}
	_, err2 := svc.Register(ctx, input2)
	if err2 == nil {
		t.Fatal("expected conflict error on duplicate email, got nil")
	}
	if !shared.IsConflict(err2) {
		t.Errorf("expected conflict error, got: %v", err2)
	}
}

// 4. Customer profile creation
func TestRegister_CustomerProfileCreation(t *testing.T) {
	svc, _, cRepo, _, _, _ := setupAuthTestEnv()
	ctx := context.Background()

	input := authmodel.RegisterInput{
		Phone:    "+919111222333",
		Password: "Password123",
		// Email omitted
	}

	out, err := svc.Register(ctx, input)
	if err != nil {
		t.Fatalf("registration without email failed: %v", err)
	}

	prof, err := cRepo.GetProfileByUserID(ctx, out.User.ID)
	if err != nil {
		t.Fatalf("customer profile should exist: %v", err)
	}
	if prof.UserID != out.User.ID {
		t.Errorf("profile user_id expected %s, got %s", out.User.ID, prof.UserID)
	}
	if string(prof.ContactPreferences) != "{}" {
		t.Errorf("expected default contact preferences {}, got %s", string(prof.ContactPreferences))
	}
}

// 5. Registration rollback when a database operation fails
func TestRegister_RollbackOnFailure(t *testing.T) {
	svc, uRepo, cRepo, txMgr, _, _ := setupAuthTestEnv()
	ctx := context.Background()

	// Simulate failure when creating customer profile
	cRepo.createProfileErr = errors.New("simulated profile creation db failure")

	input := authmodel.RegisterInput{
		Phone:    "+919555666777",
		Email:    "rollback@example.com",
		Password: "Password123",
	}

	out, err := svc.Register(ctx, input)
	if err == nil {
		t.Fatal("expected registration to fail when profile creation fails, got nil")
	}
	if out != nil {
		t.Error("expected nil output on registration failure")
	}

	if !txMgr.rolledBack {
		t.Error("expected transaction rollback to be executed")
	}

	// Verify user was NOT persisted in the repository due to rollback
	user, err := uRepo.GetUserByPhone(ctx, input.Phone)
	if !errors.Is(err, shared.ErrNotFound) || user != nil {
		t.Errorf("user should have been rolled back, but was found: %v", user)
	}

	// Verify customer profile was NOT persisted
	if len(cRepo.profiles) != 0 {
		t.Errorf("no customer profiles should exist after rollback, got %d", len(cRepo.profiles))
	}
}

// 6. Successful login
func TestLogin_Success(t *testing.T) {
	svc, _, _, _, _, _ := setupAuthTestEnv()
	ctx := context.Background()

	phone := "+919988776655"
	password := "CorrectPassword123"

	// Register user
	_, err := svc.Register(ctx, authmodel.RegisterInput{
		Phone:    phone,
		Password: password,
	})
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	// Login with correct credentials
	loginOut, err := svc.Login(ctx, authmodel.LoginInput{
		Phone:    phone,
		Password: password,
	})
	if err != nil {
		t.Fatalf("expected successful login, got: %v", err)
	}
	if loginOut == nil || loginOut.AccessToken == "" || loginOut.RefreshToken == "" {
		t.Fatal("expected valid tokens returned on login")
	}
	if loginOut.User.Status != authmodel.UserStatusActive {
		t.Errorf("expected ACTIVE user status, got %s", loginOut.User.Status)
	}
}

// 7. Incorrect password
func TestLogin_IncorrectPassword(t *testing.T) {
	svc, _, _, _, _, _ := setupAuthTestEnv()
	ctx := context.Background()

	phone := "+919988776655"
	password := "CorrectPassword123"

	_, err := svc.Register(ctx, authmodel.RegisterInput{
		Phone:    phone,
		Password: password,
	})
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	// Attempt login with wrong password
	loginOut, err := svc.Login(ctx, authmodel.LoginInput{
		Phone:    phone,
		Password: "WrongPassword999",
	})
	if err == nil {
		t.Fatal("expected login to fail with wrong password, got nil")
	}
	if loginOut != nil {
		t.Error("expected nil output on failed login")
	}

	var appErr *shared.AppError
	if !errors.As(err, &appErr) || appErr.Code != "INVALID_CREDENTIALS" {
		t.Errorf("expected INVALID_CREDENTIALS error, got: %v", err)
	}
}

// 8. Inactive user login
func TestLogin_InactiveUser(t *testing.T) {
	svc, uRepo, _, _, _, _ := setupAuthTestEnv()
	ctx := context.Background()

	phone := "+919988776644"
	password := "Password123"

	// Create user with PENDING status directly in repo
	hash, _ := auth.HashPassword(password, 64*1024, 3, 4, 16, 32)
	pendingUser := &authmodel.User{
		ID:           uuid.New(),
		Phone:        phone,
		PasswordHash: hash,
		Status:       authmodel.UserStatusPending,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	uRepo.users[pendingUser.ID] = pendingUser
	uRepo.usersByPhone[phone] = pendingUser

	// Attempt login
	_, err := svc.Login(ctx, authmodel.LoginInput{
		Phone:    phone,
		Password: password,
	})
	if err == nil {
		t.Fatal("expected login to fail for PENDING user, got nil")
	}
	var appErr *shared.AppError
	if !errors.As(err, &appErr) || appErr.Code != "ACCOUNT_INACTIVE" {
		t.Errorf("expected ACCOUNT_INACTIVE error for PENDING user, got: %v", err)
	}

	// Change status to SUSPENDED
	pendingUser.Status = authmodel.UserStatusSuspended
	_, errSuspended := svc.Login(ctx, authmodel.LoginInput{
		Phone:    phone,
		Password: password,
	})
	if errSuspended == nil {
		t.Fatal("expected login to fail for SUSPENDED user, got nil")
	}
	if !errors.As(errSuspended, &appErr) || appErr.Code != "ACCOUNT_INACTIVE" {
		t.Errorf("expected ACCOUNT_INACTIVE error for SUSPENDED user, got: %v", errSuspended)
	}
}

// 9. Refresh-token handling
func TestRefresh_Handling(t *testing.T) {
	svc, uRepo, _, _, _, clk := setupAuthTestEnv()
	ctx := context.Background()

	phone := "+919000111222"
	password := "Password123"

	regOut, err := svc.Register(ctx, authmodel.RegisterInput{
		Phone:    phone,
		Password: password,
	})
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	origRefreshToken := regOut.RefreshToken

	// 9a. Successful refresh for active user
	refreshOut, err := svc.Refresh(ctx, authmodel.RefreshInput{RefreshToken: origRefreshToken})
	if err != nil {
		t.Fatalf("expected successful refresh, got: %v", err)
	}
	if refreshOut == nil || refreshOut.AccessToken == "" || refreshOut.RefreshToken == "" {
		t.Fatal("expected new token pair from refresh")
	}

	// Verify old session was revoked
	oldHash := authservice_hashToken(origRefreshToken)
	oldSession, err := uRepo.GetRefreshSession(ctx, oldHash)
	if err != nil {
		t.Fatalf("could not retrieve old session: %v", err)
	}
	if oldSession.RevokedAt == nil {
		t.Error("expected old session to be revoked")
	}

	// 9b. Attempting refresh with already revoked token returns TOKEN_REVOKED
	_, errRevoked := svc.Refresh(ctx, authmodel.RefreshInput{RefreshToken: origRefreshToken})
	if errRevoked == nil {
		t.Fatal("expected error refreshing revoked token, got nil")
	}
	var appErr *shared.AppError
	if !errors.As(errRevoked, &appErr) || appErr.Code != "TOKEN_REVOKED" {
		t.Errorf("expected TOKEN_REVOKED, got: %v", errRevoked)
	}

	// 9c. Inactive user refresh attempt does NOT revoke session
	activeRefresh := refreshOut.RefreshToken
	user := uRepo.users[regOut.User.ID]
	user.Status = authmodel.UserStatusSuspended

	_, errInactive := svc.Refresh(ctx, authmodel.RefreshInput{RefreshToken: activeRefresh})
	if errInactive == nil {
		t.Fatal("expected error refreshing for inactive user, got nil")
	}
	if !errors.As(errInactive, &appErr) || appErr.Code != "ACCOUNT_INACTIVE" {
		t.Errorf("expected ACCOUNT_INACTIVE, got: %v", errInactive)
	}

	// Check that session was NOT revoked because user check happened BEFORE revocation
	activeHash := authservice_hashToken(activeRefresh)
	activeSession, _ := uRepo.GetRefreshSession(ctx, activeHash)
	if activeSession.RevokedAt != nil {
		t.Error("old session should NOT be revoked if user is inactive")
	}

	// 9d. Expired token handling
	user.Status = authmodel.UserStatusActive
	// Advance clock past expiration (refresh TTL is 24h)
	clk.now = clk.now.Add(48 * time.Hour)

	_, errExpired := svc.Refresh(ctx, authmodel.RefreshInput{RefreshToken: activeRefresh})
	if errExpired == nil {
		t.Fatal("expected error refreshing expired token, got nil")
	}
	if !errors.As(errExpired, &appErr) || appErr.Code != "TOKEN_EXPIRED" {
		t.Errorf("expected TOKEN_EXPIRED, got: %v", errExpired)
	}
}

// Helper to hash token matching service hashToken logic
func authservice_hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
