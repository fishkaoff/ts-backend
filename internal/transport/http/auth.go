package httpserver

import (
	"errors"
	"net/http"

	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/services/auth"
	"github.com/fishkaoff/ts-backend/internal/transport"
	"github.com/go-chi/chi/v5"
)

func (s *HTTPServer) AuthRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", s.HandleRegister)
	r.Post("/login", s.HandleLogin)

	return r
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

	tokens, err := s.authService.Login(r.Context(), dto)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			writeError(w, *transport.ApiBadRequest.AddDetails(err.Error()))
			return
		}

		writeError(w, *transport.InternalError.AddDetails("try again later"))
		return
	}

	writeJSON(w, 200, tokens)
}
