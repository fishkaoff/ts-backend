package httpserver

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/services/products"
	"github.com/fishkaoff/ts-backend/internal/transport"
	"github.com/go-chi/chi/v5"
)

func (s *HTTPServer) ProductsRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", s.HandleGetProducts)
	r.Get("/{id}", s.HandleGetProduct)

	return r
}

func (s *HTTPServer) HandleGetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var filter types.ProductsFilter

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		writeError(w, *transport.ApiBadRequest.AddDetails(err.Error()))
		return
	}

	if limit <= 0 {
		writeError(w, *transport.ApiBadRequest.AddDetails("Лимит должен быть больше нуля"))
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		writeError(w, *transport.ApiBadRequest.AddDetails(err.Error()))
		return
	}

	if offset < 0 {
		writeError(w, *transport.ApiBadRequest.AddDetails("Сдвиг должен быть больше нуля"))
		return
	}

	filter.Offset = offset
	filter.Limit = limit

	products, err := s.productsService.GetProducts(ctx, filter)
	if err != nil {
		writeError(w, transport.InternalError)
		return
	}

	writeJSON(w, 200, products)
}

func (s *HTTPServer) HandleGetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	productId := chi.URLParam(r, "id")

	product, err := s.productsService.GetProductById(ctx, productId)
	if err != nil {
		if errors.Is(err, products.ErrProductNotFound) {
			writeError(w, *transport.ApiBadRequest.AddDetails("Товар с таким id не найден"))
			return
		}

		writeError(w, transport.InternalError)
		return
	}

	writeJSON(w, 200, product)
}
