package products

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"time"

	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/storage"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrProductNotFound = errors.New("Товар не найден")

type ProductsStore interface {
	UpdateProducts(ctx context.Context, products []types.Product) error
	GetProductById(ctx context.Context, id string) (types.Product, error)
	GetProducts(ctx context.Context, filter types.ProductsFilter) ([]types.Product, error)
	GetProductsByIds(ctx context.Context, ids []bson.ObjectID) ([]types.Product, error)
}

type ProductsAdapter interface {
	ParseProductsFromFile(ctx context.Context, file multipart.File) ([]types.Product, error)
}

type ProductsService struct {
	log             *slog.Logger
	store           ProductsStore
	productsAdapter ProductsAdapter
}

func New(log *slog.Logger, store ProductsStore, productsAdapter ProductsAdapter) *ProductsService {
	return &ProductsService{
		log:             log,
		store:           store,
		productsAdapter: productsAdapter,
	}
}

func (s *ProductsService) GetProducts(ctx context.Context, filter types.ProductsFilter) ([]types.Product, error) {
	const op = "products.GetProducts"
	log := s.log.With("op", op)

	log.Info("get products", "filter", filter)
	products, err := s.store.GetProducts(ctx, filter)
	if err != nil {
		log.Error("%s:%w", op, err)
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("got products")
	return products, nil
}

func (s *ProductsService) GetProductById(ctx context.Context, id string) (types.Product, error) {
	const op = "products.GetProductById"
	log := s.log.With("op", op)

	log.Info("get product by id")
	product, err := s.store.GetProductById(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrDocumentNotFound) {
			log.Info("product not found")
			return types.Product{}, ErrProductNotFound
		}

		log.Error(fmt.Errorf("%s: %w", op, err).Error())
		return types.Product{}, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("product found")

	return product, nil
}

func (s *ProductsService) GetProductsByIds(ctx context.Context, ids []bson.ObjectID) ([]types.Product, error) {
	const op = "products.GetProductsByIds"
	log := s.log.With("op", op)

	log.Info("get products by ids", "ids", ids)
	products, err := s.store.GetProductsByIds(ctx, ids)
	if err != nil {
		log.Error("%s:%w", op, err)
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("got products by ids")
	return products, nil
}

func (s *ProductsService) UpdatePrice(ctx context.Context, file multipart.File) error {
	const op = "products.UpdatePrice"
	log := s.log.With("op", op)

	log.Info("update price")

	log.Info("parse file")
	products, err := s.productsAdapter.ParseProductsFromFile(ctx, file)
	if err != nil {
		log.Error("error while parsing", "error", err)
		return fmt.Errorf("%s:%w", op, err)
	}

	go func(products []types.Product) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		log.Info("start saving products in background")
		if err := s.saveProducts(bgCtx, products); err != nil {
			log.Error("error while saving products", "error", err)
		} else {
			log.Info("products saved successfully")
		}
		cancel()
	}(products)

	log.Info("price update started in background")

	return nil
}

func (s *ProductsService) saveProducts(ctx context.Context, products []types.Product) error {
	const op = "products.saveProducts"

	err := s.store.UpdateProducts(ctx, products)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}
