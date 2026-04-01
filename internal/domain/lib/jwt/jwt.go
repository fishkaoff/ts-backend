package jwtclient

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	Role string `json:"role"`
	Id   string `json:"user_id"`

	jwt.RegisteredClaims
}

type Service struct {
	secretKey []byte
}

func New(secretKey []byte) *Service {
	return &Service{
		secretKey: secretKey,
	}
}

func (s *Service) IsValidJWT(tokenString string) (bool, error) {
	token, _, err := s.parseToken(tokenString)
	if err != nil {
		return false, err
	}

	if !token.Valid {
		return false, errors.New("invalid token")
	}

	return true, nil
}

func (s *Service) ExtractClaims(tokenString string) (*CustomClaims, error) {
	token, claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *Service) parseToken(tokenString string) (*jwt.Token, *CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.secretKey, nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return token, claims, nil
}
