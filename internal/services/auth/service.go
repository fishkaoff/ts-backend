package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/fishkaoff/ts-backend/internal/config"
	jwtclient "github.com/fishkaoff/ts-backend/internal/domain/lib/jwt"
	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserAlreadyExists = errors.New("Пользователь с такиим email уже существует")
var ErrUserNotFound = errors.New("Пользователь с таким email не найден")
var ErrIncorrectPassword = errors.New("Неверный пароль")

type AuthStore interface {
	SaveUser(ctx context.Context, user types.User) (types.User, error)
	GetUserByEmail(ctx context.Context, email string) (types.User, error)
	GetUserById(ctx context.Context, id string) (types.User, error)
	DeleteUserByEmail(ctx context.Context, email string) error
}

type Service struct {
	log    *slog.Logger
	jwtCfg config.JWTConfig
	store  AuthStore
}

func New(log *slog.Logger, jwtCfg config.JWTConfig, store AuthStore) *Service {
	return &Service{
		log:    log,
		jwtCfg: jwtCfg,
		store:  store,
	}
}

func (s *Service) RegisterUser(ctx context.Context, dto types.RegisterDto) (types.User, error) {
	const op = "authservice.RegisterUser"

	logger := s.log.With("op", op)
	logger.Info("registering user")

	exists, err := s.IsUserExists(ctx, dto.Email)
	if err != nil {
		logger.Error(fmt.Errorf("%s:%w", op, err).Error())
		return types.User{}, nil
	}

	if exists {
		logger.Info("user already exists")
		return types.User{}, ErrUserAlreadyExists
	}

	passHash, err := generatePassHash(dto.Password)
	if err != nil {
		return types.User{}, fmt.Errorf("%s:%w", op, err)
	}

	user := types.User{
		Email:    dto.Email,
		Name:     dto.Name,
		Lastname: dto.Lastname,
		PassHash: passHash,
		Role:     types.USER,
	}

	savedUser, err := s.store.SaveUser(ctx, user)
	if err != nil {
		logger.Error(fmt.Errorf("%s:%w", op, err).Error())
		return types.User{}, fmt.Errorf("%s:%w", op, err)
	}

	logger.Info("user registered")
	return savedUser, nil
}

func (s *Service) Login(ctx context.Context, dto types.LoginDto) (types.User, types.TokenPair, error) {
	const op = "authservice.Login"

	logger := s.log.With("op", op)
	logger.Info("login user")

	user, err := s.store.GetUserByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			logger.Info("user not found")
			return types.User{}, types.TokenPair{}, ErrUserNotFound
		}

		logger.Error(fmt.Errorf("%s:%w", op, err).Error())
		return types.User{}, types.TokenPair{}, fmt.Errorf("%s:%w", op, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(dto.Password))
	if err != nil {
		logger.Info("incorrect password")
		return types.User{}, types.TokenPair{}, ErrIncorrectPassword
	}

	tokens, err := generateTokenPair(user, s.jwtCfg)
	if err != nil {
		logger.Error(fmt.Errorf("%s:%w", op, err).Error())
		return types.User{}, types.TokenPair{}, fmt.Errorf("%s:%w", op, err)
	}

	return user, tokens, nil

}

func (s *Service) GetUserInfo(ctx context.Context, id string) (types.User, error) {
	const op = "authservice.GetUserInfo"

	user, err := s.store.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return types.User{}, ErrUserNotFound
		}

		return types.User{}, fmt.Errorf("%s:%w", op, err)
	}

	return user, nil
}

func (s *Service) IsUserExists(ctx context.Context, email string) (bool, error) {
	const op = "authservice.IsUserExists"

	_, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return false, nil
		}

		return true, fmt.Errorf("%s:%w", op, err)
	}

	return true, nil
}

func generatePassHash(password string) ([]byte, error) {
	const op = "authservice.generatePassHash"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return hash, nil
}

func generateTokenPair(
	user types.User,
	jwtConfig config.JWTConfig,
) (types.TokenPair, error) {
	const op = "authservice.generateTokenPair"

	accessClaims := jwtclient.CustomClaims{
		Id:   user.Id.Hex(),
		Role: parseRole(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(time.Hour * time.Duration(jwtConfig.AccessTokenLifeTime)),
			),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	refreshClaims := jwtclient.CustomClaims{
		Id:   user.Id.Hex(),
		Role: parseRole(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(time.Hour * time.Duration(jwtConfig.RefreshTokenLifeTime)),
			),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		accessClaims,
	).SignedString([]byte(jwtConfig.SecretKey))
	if err != nil {
		return types.TokenPair{}, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		refreshClaims,
	).SignedString([]byte(jwtConfig.SecretKey))
	if err != nil {
		return types.TokenPair{}, fmt.Errorf("%s: %w", op, err)
	}

	return types.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func parseRole(role types.UserRole) string {
	switch role {
	case "admin":
		return "admin"
	default:
		return "admin"
	}
}
