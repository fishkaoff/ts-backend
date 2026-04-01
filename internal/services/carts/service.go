package carts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/storage"
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

type CartsService struct {
	log        *slog.Logger
	cartsStore CartsStore
}

func New(log *slog.Logger, cartsStore CartsStore) *CartsService {
	return &CartsService{
		log:        log,
		cartsStore: cartsStore,
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
