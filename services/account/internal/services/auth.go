package services

import (
	"context"
	"errors"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/gianglt2198/federation-go/package/helpers"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"

	"github.com/gianglt2198/federation-go/services/account/generated/ent"
	"github.com/gianglt2198/federation-go/services/account/generated/ent/user"
	"github.com/gianglt2198/federation-go/services/account/generated/graph/model"
	"github.com/gianglt2198/federation-go/services/account/internal/repos"
)

type (
	authService struct {
		log *monitoring.Logger

		userRepository    repos.UserRepository
		sessionRepository repos.SessionRepository
		jwtHelper         helpers.JwtHelper
		encryptor         helpers.Encryptor
	}

	AuthService interface {
		Register(ctx context.Context, input model.RegisterInput) (*ent.User, error)
		Login(ctx context.Context, input model.LoginInput) (string, error)
		Logout(ctx context.Context, sessionID pnnid.ID) error
	}
)

type AuthServiceParams struct {
	fx.In

	Log *monitoring.Logger

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
			userRepository:    params.UserRepository,
			sessionRepository: params.SessionRepository,
			jwtHelper:         params.JWTHelper,
			encryptor:         params.Encryptor,
		},
	}
}

func (s *authService) Register(ctx context.Context, input model.RegisterInput) (*ent.User, error) {
	s.log.InfoC(ctx, "Starting user registration", zap.String("username", input.Username))

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
		FirstName: getStringOrDefault(input.FirstName, ""),
		LastName:  getStringOrDefault(input.LastName, ""),
		Phone:     getStringOrDefault(input.Phone, ""),
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

	// Generate JWT token
	claims := map[string]interface{}{
		"user_id":  string(foundUser.ID),
		"username": foundUser.Username,
		"email":    foundUser.Email,
	}

	token, err := s.jwtHelper.GenerateToken(claims)
	if err != nil {
		s.log.ErrorC(ctx, "Failed to generate JWT token", zap.Error(err))
		return "", errors.New("failed to generate authentication token")
	}

	// Create session record
	sessionInput := ent.CreateSessionInput{
		UserID:     string(foundUser.ID),
		LastUsedAt: time.Now(),
	}

	session, err := s.sessionRepository.CreateOne(ctx, sessionInput)
	if err != nil {
		s.log.ErrorC(ctx, "Failed to create session", zap.Error(err))
		// Don't fail login if session creation fails, just log it
	} else {
		s.log.InfoC(ctx, "Session created",
			zap.String("session_id", string(session.ID)),
			zap.String("user_id", string(foundUser.ID)),
		)
	}

	s.log.InfoC(ctx, "User logged in successfully",
		zap.String("user_id", string(foundUser.ID)),
		zap.String("username", foundUser.Username),
	)

	return token, nil
}

func (s *authService) Logout(ctx context.Context, sessionID pnnid.ID) error {
	s.log.InfoC(ctx, "Starting user logout", zap.String("session_id", string(sessionID)))

	err := s.sessionRepository.DeleteOne(ctx, sessionID, nil)
	if err != nil {
		s.log.ErrorC(ctx, "Failed to delete session", zap.Error(err))
		return errors.New("failed to logout")
	}

	s.log.InfoC(ctx, "User logged out successfully", zap.String("session_id", string(sessionID)))
	return nil
}

// Helper function to get string value or default
func getStringOrDefault(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}
