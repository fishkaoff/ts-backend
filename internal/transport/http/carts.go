package httpserver

import (
	"net/http"

	"github.com/fishkaoff/ts-backend/internal/domain/lib/utils"
	"github.com/fishkaoff/ts-backend/internal/transport"
	"github.com/fishkaoff/ts-backend/internal/transport/middlewares"
	"github.com/go-chi/chi/v5"
)

func (s *HTTPServer) CartsRoutes(jwtMidlleware *middlewares.JWTMiddleware) chi.Router {
	r := chi.NewRouter()

	r.With(jwtMidlleware.Middleware).Get("/", s.GetUsersCart)
	r.With(jwtMidlleware.Middleware).Put("/items", s.UpdateProductQuantity)

	return r
}

func (s *HTTPServer) GetUsersCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := utils.ExtractUserIdFromCtx(ctx)
	if err != nil {
		writeError(w, *transport.ApiBadRequest.AddDetails(err.Error()))
		return
	}

	cart, err := s.cartsService.GetUsersCartFull(ctx, userId)
	if err != nil {
		writeError(w, *transport.InternalError.AddDetails(err.Error()))
		return
	}

	writeJSON(w, 200, cart)
}

func (s *HTTPServer) UpdateProductQuantity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := utils.ExtractUserIdFromCtx(ctx)
	if err != nil {
		writeError(w, *transport.ApiBadRequest.AddDetails(err.Error()))
		return
	}

	var dto struct {
		ProductId string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}

	err = parseBody(r, &dto)
	if err != nil {
		writeError(w, *transport.ApiBadRequest.AddDetails("Невалидное тело запроса"))
		return
	}

	if dto.Quantity < 0 {
		writeError(w, *transport.ApiBadRequest.AddDetails("Количество должно быть 0 или больше"))
		return
	}

	err = s.cartsService.UpdateProductQuantity(ctx, userId, dto.ProductId, dto.Quantity)
	if err != nil {
		writeError(w, *transport.InternalError.AddDetails(err.Error()))
		return
	}

	writeJSON(w, 200, map[string]bool{
		"success": true,
	})
}
