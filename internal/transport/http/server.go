package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"

	"github.com/fishkaoff/ts-backend/internal/config"
	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/transport"
	"github.com/fishkaoff/ts-backend/internal/transport/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type AuthService interface {
	RegisterUser(ctx context.Context, dto types.RegisterDto) (types.User, error)
	Login(ctx context.Context, dto types.LoginDto) (types.User, types.TokenPair, error)
	GetUserInfo(ctx context.Context, id string) (types.User, error)
}

type ProductsService interface {
	UpdatePrice(ctx context.Context, file multipart.File) error
	GetProducts(ctx context.Context, filter types.ProductsFilter) ([]types.Product, error)
	GetProductById(ctx context.Context, id string) (types.Product, error)
}

type CartsService interface {
	GetUsersCart(ctx context.Context, userId string) (types.Cart, error)
	UpdateProductQuantity(
		ctx context.Context, userId string,
		productId string,
		quantity int,
	) error
	GetUsersCartFull(
		ctx context.Context,
		userId string,
	) (types.CartFull, error)
}

type HTTPServer struct {
	log       *slog.Logger
	cfg       config.RESTConfig
	jwtClient middlewares.JwtChecker

	authService     AuthService
	productsService ProductsService
	cartsService    CartsService
}

func New(
	cfg config.RESTConfig,
	log *slog.Logger,
	jwtClient middlewares.JwtChecker,
	authSvc AuthService,
	productsSvc ProductsService,
	cartsService CartsService,
) *HTTPServer {
	return &HTTPServer{
		cfg:             cfg,
		log:             log,
		jwtClient:       jwtClient,
		authService:     authSvc,
		productsService: productsSvc,
		cartsService:    cartsService,
	}
}

func (s *HTTPServer) Start() error {
	r := s.setupRoutes(chi.NewRouter())
	return http.ListenAndServe(s.cfg.Addr, r)
}

func (s *HTTPServer) setupRoutes(mainRouter *chi.Mux) *chi.Mux {
	mainRouter.Use(middleware.RequestID)
	mainRouter.Use(middleware.RealIP)
	mainRouter.Use(middleware.Logger)
	mainRouter.Use(middleware.Recoverer)

	jwt := middlewares.NewJWTMiddleware(s.jwtClient)

	mainRouter.Route("/api", func(api chi.Router) {
		api.Mount("/auth", s.AuthRoutes(jwt))
		api.Mount("/products", s.ProductsRoutes())
		api.Mount("/admin", s.AdminRoutes())
		api.Mount("/carts", s.CartsRoutes(jwt))
	})

	return mainRouter
}

func parseBody[T any](r *http.Request, target *T) error {
	if r.Body != nil && r.ContentLength > 0 {
		defer r.Body.Close()

		err := json.NewDecoder(r.Body).Decode(&target)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}

	return nil
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, err transport.ApiError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)

	_ = json.NewEncoder(w).Encode(err)
}
