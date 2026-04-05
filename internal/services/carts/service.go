package carts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/storage"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrInvalidQuantity = errors.New("Количество не может быть меньше нуля")

type CartsStore interface {
	GetCartByUserId(ctx context.Context, userId string) (types.Cart, error)
	UpsertProduct(
		ctx context.Context,
		userId string,
		productId string,
		quantity int,
	) error
}

type ProductsProvider interface {
	GetProductsByIds(ctx context.Context, ids []bson.ObjectID) ([]types.Product, error)
}

type CartsService struct {
	log              *slog.Logger
	cartsStore       CartsStore
	productsProvider ProductsProvider
}

func New(log *slog.Logger, cartsStore CartsStore, productsProvider ProductsProvider) *CartsService {
	return &CartsService{
		log:              log,
		cartsStore:       cartsStore,
		productsProvider: productsProvider,
	}
}

func (s *CartsService) GetUsersCart(ctx context.Context, userId string) (types.Cart, error) {
	const op = "carts.GetUsersCart"
	log := s.log.With("op", op)

	log.Info("get users cart")
	cart, err := s.cartsStore.GetCartByUserId(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrBadUserId) {
			log.Error("invalid user id", "id", userId)
			return types.Cart{}, fmt.Errorf("%s:%w", op, err)
		}

		log.Error(fmt.Errorf("%s:%w", op, err).Error())
		return types.Cart{}, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("got users cart")
	return cart, nil
}

func (s *CartsService) UpdateProductQuantity(
	ctx context.Context,
	userId string,
	productId string,
	quantity int,
) error {
	const op = "carts.UpdateProductQuantity"
	log := s.log.With("op", op)

	log.Info("update product quantity",
		"quantity", quantity,
		"product", productId,
	)

	if quantity < 0 {
		log.Info("invalid product quantity")
		return ErrInvalidQuantity
	}

	err := s.cartsStore.UpsertProduct(ctx, userId, productId, quantity)
	if err != nil {
		if errors.Is(err, storage.ErrBadUserId) {
			log.Error("invalid user id", "id", userId)
			return fmt.Errorf("%s:%w", op, err)
		}

		if errors.Is(err, storage.ErrBadProductId) {
			log.Error("bad product id", "product_id", productId)
			return fmt.Errorf("%s:%w", op, err)
		}

		log.Error(fmt.Errorf("%s:%w", op, err).Error())
		return fmt.Errorf("%s:%w", op, err)
	}

	log.Info("updated product quantity")
	return nil
}

func (s *CartsService) GetUsersCartFull(
	ctx context.Context,
	userId string,
) (types.CartFull, error) {
	const op = "carts.GetUsersCartFull"
	log := s.log.With("op", op)

	log.Info("get users cart full")

	// 1. получаем корзину (с product_id)
	cart, err := s.cartsStore.GetCartByUserId(ctx, userId)
	if err != nil {
		log.Error("failed to get cart", "error", err)
		return types.CartFull{}, fmt.Errorf("%s:%w", op, err)
	}

	// если корзина пустая — сразу возвращаем
	if len(cart.Products) == 0 {
		return types.CartFull{
			Id:       cart.Id,
			UserId:   cart.UserId,
			Products: []types.CartItemFull{},
		}, nil
	}

	// собираем product IDs
	ids := make([]bson.ObjectID, 0, len(cart.Products))
	for _, item := range cart.Products {
		ids = append(ids, item.ProductId)
	}

	// получаем продукты пачкой
	products, err := s.productsProvider.GetProductsByIds(ctx, ids)
	if err != nil {
		log.Error("failed to get products", "error", err)
		return types.CartFull{}, fmt.Errorf("%s:%w", op, err)
	}

	// делаем map для быстрого поиска
	productMap := make(map[bson.ObjectID]types.Product, len(products))
	for _, p := range products {
		productMap[p.Id] = p
	}

	// собираем итоговый ответ
	result := types.CartFull{
		Id:       cart.Id,
		UserId:   cart.UserId,
		Products: make([]types.CartItemFull, 0, len(cart.Products)),
	}

	for _, item := range cart.Products {
		product, ok := productMap[item.ProductId]
		if !ok {
			log.Warn("product not found", "product_id", item.ProductId.Hex())
			continue
		}

		result.Products = append(result.Products, types.CartItemFull{
			Product:  product,
			Quantity: item.Quantity,
		})
	}

	log.Info("got users cart full")

	return result, nil
}
