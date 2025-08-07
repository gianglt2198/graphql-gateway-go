package services

import (
	"context"
	"errors"
	"time"

	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/gianglt2198/federation-go/package/helpers"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/utils"

	"github.com/gianglt2198/federation-go/services/account/config"
	"github.com/gianglt2198/federation-go/services/account/generated/ent"
	"github.com/gianglt2198/federation-go/services/account/generated/ent/session"
	"github.com/gianglt2198/federation-go/services/account/generated/ent/user"
	"github.com/gianglt2198/federation-go/services/account/generated/graph/model"
	"github.com/gianglt2198/federation-go/services/account/internal/repos"
)

type (
	authService struct {
		log *monitoring.Logger

		config config.AccountConfig

		jwtHelper helpers.JwtHelper
		encryptor helpers.Encryptor

		userRepository    repos.UserRepository
		sessionRepository repos.SessionRepository
	}

	AuthService interface {
		FindAuthByID(ctx context.Context, id string) (*model.AuthVerifyEntity, error)
		Register(ctx context.Context, input model.RegisterInput) (*ent.User, error)
		Login(ctx context.Context, input model.LoginInput) (string, error)
		Logout(ctx context.Context) error
		AuthVerify(ctx context.Context, input model.AuthVerifyInput) (*model.AuthVerifyEntity, error)
	}
)

type AuthServiceParams struct {
	fx.In

	Log *monitoring.Logger

	Config config.AccountConfig

	JWTHelper helpers.JwtHelper
	Encryptor helpers.Encryptor

	UserRepository    repos.UserRepository
	SessionRepository repos.SessionRepository
}

type AuthServiceResult struct {
	fx.Out

	AuthService AuthService
}

func NewAuthService(params AuthServiceParams) AuthServiceResult {
	return AuthServiceResult{
		AuthService: &authService{
			log:               params.Log,
			config:            params.Config,
			jwtHelper:         params.JWTHelper,
			encryptor:         params.Encryptor,
			userRepository:    params.UserRepository,
			sessionRepository: params.SessionRepository,
		},
	}
}

func (s *authService) Register(ctx context.Context, input model.RegisterInput) (*ent.User, error) {
	s.log.InfoC(ctx, "Starting user registration", zap.String("username", input.Username))

	ctx = utils.ApplyUserIDWithContext(ctx, "system")

	// Check if user already exists
	existingUser, err := s.userRepository.Query(ctx).
		Where(user.Or(user.UsernameEQ(input.Username), user.EmailEQ(input.Email))).
		First(ctx)
	if err == nil && existingUser != nil {
		s.log.WarnC(ctx, "User already exists", zap.String("username", input.Username))
		return nil, errors.New("user already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.ErrorC(ctx, "Failed to hash password", zap.Error(err))
		return nil, errors.New("failed to process password")
	}

	// Create user input
	createUserInput := ent.CreateUserInput{
		Username:  input.Username,
		Email:     input.Email,
		Password:  string(hashedPassword),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Phone:     input.Phone,
	}

	// Create the user
	newUser, err := s.userRepository.CreateOne(ctx, createUserInput)
	if err != nil {
		s.log.ErrorC(ctx, "Failed to create user", zap.Error(err))
		return nil, errors.New("failed to create user")
	}

	s.log.InfoC(ctx, "User registered successfully",
		zap.String("user_id", string(newUser.ID)),
		zap.String("username", newUser.Username),
	)

	return newUser, nil
}

func (s *authService) Login(ctx context.Context, input model.LoginInput) (string, error) {
	s.log.InfoC(ctx, "Starting user login", zap.String("username", input.Username))

	// Find user by username
	foundUser, err := s.userRepository.Query(ctx).
		Where(user.UsernameEQ(input.Username)).
		First(ctx)
	if err != nil {
		s.log.WarnC(ctx, "User not found", zap.String("username", input.Username))
		return "", errors.New("invalid username or password")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(input.Password))
	if err != nil {
		s.log.WarnC(ctx, "Invalid password", zap.String("username", input.Username))
		return "", errors.New("invalid username or password")
	}

	_, err = s.sessionRepository.DeleteWithPredicates(ctx, session.HasUserWith(user.IDEQ(string(foundUser.ID))))
	if err != nil {
		s.log.ErrorC(ctx, "Failed to delete sessions", zap.Error(err))
		return "", errors.New("failed to delete sessions")
	}

	// Create session record
	sessionInput := ent.CreateSessionInput{
		UserID:     string(foundUser.ID),
		LastUsedAt: time.Now(),
	}

	sess, err := s.sessionRepository.CreateOne(ctx, sessionInput)
	if err != nil {
		s.log.ErrorC(ctx, "Failed to create session", zap.Error(err))
		return "", errors.New("failed to create session")
	}

	// Generate JWT token
	claims := map[string]interface{}{
		"iat":      time.Now().Unix(),
		"sub":      sess.ID,
		"user_id":  string(foundUser.ID),
		"username": foundUser.Username,
		"email":    foundUser.Email,
	}

	token, err := s.jwtHelper.GenerateToken(claims)
	if err != nil {
		s.log.ErrorC(ctx, "Failed to generate JWT token", zap.Error(err))
		return "", errors.New("failed to generate authentication token")
	}

	s.log.InfoC(ctx, "User logged in successfully",
		zap.String("user_id", string(foundUser.ID)),
		zap.String("username", foundUser.Username),
	)

	return token, nil
}

func (s *authService) Logout(ctx context.Context) error {
	userID := ctx.Value("user_id").(string)

	_, err := s.sessionRepository.DeleteWithPredicates(ctx, session.HasUserWith(user.IDEQ(userID)))
	if err != nil {
		s.log.ErrorC(ctx, "Failed to delete session", zap.Error(err))
		return errors.New("failed to logout")
	}

	return nil
}

func (s *authService) FindAuthByID(ctx context.Context, id string) (*model.AuthVerifyEntity, error) {
	session, err := s.sessionRepository.Query(ctx).
		Where(session.IDEQ(string(id))).
		WithUser().
		First(ctx)
	if err != nil {
		return nil, err
	}

	sessionExpired := session.LastUsedAt.Add(time.Duration(s.config.ExpiredDuration) * time.Hour)

	return &model.AuthVerifyEntity{
		ID:               session.ID,
		UserID:           session.Edges.User.ID,
		SessionExpiredAt: sessionExpired,
		LastUsedAt:       session.LastUsedAt,
	}, nil
}

func (s *authService) AuthVerify(ctx context.Context, input model.AuthVerifyInput) (*model.AuthVerifyEntity, error) {
	claims, err := s.jwtHelper.ValidateToken(input.Token)
	if err != nil {
		return nil, err
	}

	sessionID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("session_id not found")
	}

	session, err := s.sessionRepository.Query(ctx).
		Where(session.IDEQ(sessionID)).
		WithUser().
		First(ctx)
	if err != nil {
		return nil, err
	}

	if session.LastUsedAt.Add(time.Duration(s.config.ExpiredDuration) * time.Hour).Before(time.Now()) {
		return nil, errors.New("session expired")
	}

	lastUsedAt := time.Now()

	_, err = s.sessionRepository.UpdateOne(ctx, sessionID, ent.UpdateSessionInput{
		LastUsedAt: lo.ToPtr(lastUsedAt),
	})
	if err != nil {
		return nil, err
	}

	return &model.AuthVerifyEntity{
		ID:               session.ID,
		UserID:           session.Edges.User.ID,
		SessionExpiredAt: lastUsedAt.Add(time.Duration(s.config.ExpiredDuration) * time.Hour),
		LastUsedAt:       lastUsedAt,
	}, nil
}
