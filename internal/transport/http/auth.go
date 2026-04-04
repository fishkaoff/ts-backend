package httpserver

import (
	"errors"
	"net/http"

	"github.com/fishkaoff/ts-backend/internal/domain/lib/utils"
	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/services/auth"
	"github.com/fishkaoff/ts-backend/internal/transport"
	"github.com/fishkaoff/ts-backend/internal/transport/middlewares"
	"github.com/go-chi/chi/v5"
)

func (s *HTTPServer) AuthRoutes(jwtMidlleware *middlewares.JWTMiddleware) chi.Router {
	r := chi.NewRouter()

	r.Post("/register", s.HandleRegister)
	r.Post("/login", s.HandleLogin)
	r.With(jwtMidlleware.Middleware).Get("/users", s.HandleGetUserInfo)

	return r
}

func (s *HTTPServer) HandleGetUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := utils.ExtractUserIdFromCtx(ctx)
	if err != nil {
		writeError(w, *transport.ApiBadRequest.AddDetails(err.Error()))
		return
	}

	user, err := s.authService.GetUserInfo(ctx, userId)
	if err != nil {
		writeError(w, *transport.InternalError.AddDetails(err.Error()))
		return
	}

	writeJSON(w, 200, user)
}

func (s *HTTPServer) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var dto types.RegisterDto

	err := parseBody(r, &dto)
	if err != nil {
		writeError(w, *transport.ApiBadRequest.AddDetails(err.Error()))
		return
	}

	user, err := s.authService.RegisterUser(r.Context(), dto)
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			writeError(w, *transport.ApiBadRequest.AddDetails(err.Error()))
			return
		}

		writeError(w, *transport.InternalError.AddDetails("try again later"))
		return
	}

	writeJSON(w, 200, user)
}

func (s *HTTPServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var dto types.LoginDto

	err := parseBody(r, &dto)
	if err != nil {
		writeError(w, *transport.ApiBadRequest.AddDetails("invalid body"))
		return
	}

	user, tokens, err := s.authService.Login(r.Context(), dto)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) || errors.Is(err, auth.ErrIncorrectPassword) {
			writeError(w, *transport.ApiBadRequest.AddDetails(err.Error()))
			return
		}

		writeError(w, *transport.InternalError.AddDetails("try again later"))
		return
	}

	writeJSON(w, 200, struct {
		User   interface{} `json:"user"`
		Tokens interface{} `json:"tokens"`
	}{
		User:   user,
		Tokens: tokens,
	})
}
