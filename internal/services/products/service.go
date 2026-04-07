package products

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"time"

	excel_adapter "github.com/fishkaoff/ts-backend/internal/adapters/excel"
	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/storage"
	"github.com/fishkaoff/ts-backend/internal/storage/meilisearch"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrProductNotFound = errors.New("Товар не найден")
	ErrEmptyFile       = errors.New("Файл пустой")
	ErrNoSheets        = errors.New("Не удалось найти листы в файле")
)

type ProductsStore interface {
	UpdateProducts(ctx context.Context, products []types.Product) error
	GetProductById(ctx context.Context, id string) (types.Product, error)
	GetProducts(ctx context.Context, filter types.ProductsFilter) ([]types.Product, error)
	GetProductsByIds(ctx context.Context, ids []bson.ObjectID) ([]types.Product, error)
	GetProductsByPartNumbers(ctx context.Context, partNumbers []string) ([]types.Product, error)
}

type SearchProvider interface {
	SearchProducts(query string, page int, limit int) ([]meilisearch.Product, error)
	UpsertProducts(products []types.Product) error
}

type ProductsAdapter interface {
	ParseProductsFromFile(ctx context.Context, file multipart.File) ([]types.Product, error)
}

type ProductsService struct {
	log             *slog.Logger
	store           ProductsStore
	productsAdapter ProductsAdapter
	searchProvider  SearchProvider
}

func New(log *slog.Logger, store ProductsStore, productsAdapter ProductsAdapter, searchProvider SearchProvider) *ProductsService {
	return &ProductsService{
		log:             log,
		store:           store,
		productsAdapter: productsAdapter,
		searchProvider:  searchProvider,
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

func (s *ProductsService) SearchProducts(ctx context.Context, query string, page int, limit int) ([]types.Product, error) {
	const op = "products.SearchProducts"
	log := s.log.With("op", op)

	log.Info("search products", "query", query)
	foundProducts, err := s.searchProvider.SearchProducts(query, page, limit)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	partNumbers := make([]string, 0, len(foundProducts))

	for _, product := range foundProducts {
		partNumbers = append(partNumbers, product.PartNumber)
	}

	products, err := s.store.GetProductsByPartNumbers(ctx, partNumbers)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return products, nil
}

func (s *ProductsService) UpdatePrice(ctx context.Context, file multipart.File) error {
	const op = "products.UpdatePrice"
	log := s.log.With("op", op)

	log.Info("update price")

	log.Info("parse file")
	products, err := s.productsAdapter.ParseProductsFromFile(ctx, file)
	if err != nil {
		if errors.Is(err, excel_adapter.ErrEmptyFile) {
			return ErrEmptyFile
		}

		if errors.Is(err, excel_adapter.ErrNoSheets) {
			return ErrNoSheets
		}

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

	go func(products []types.Product) {
		log.Info("start index products")

		if err := s.searchProvider.UpsertProducts(products); err != nil {
			log.Error("failed to index products", "error", err)
		} else {
			log.Info("index was created successfully")
		}
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
