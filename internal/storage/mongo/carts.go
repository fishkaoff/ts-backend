package mongostorage

import (
	"context"
	"fmt"

	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/storage"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// GetCartByUserId finds cart by user id, if not found creates it
func (s *MongoStore) GetCartByUserId(ctx context.Context, userId string) (types.Cart, error) {
	const op = "mongostorage.GetCartByUserId"

	userObjId, err := bson.ObjectIDFromHex(userId)
	if err != nil {
		return types.Cart{}, storage.ErrBadUserId
	}

	filter := bson.M{
		"user_id": userObjId,
	}

	update := bson.M{
		"$setOnInsert": bson.M{
			"user_id":  userObjId,
			"products": []types.CartItem{},
		},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var cart types.Cart
	err = s.cartsCollection.FindOneAndUpdate(
		ctx,
		filter,
		update,
		opts,
	).Decode(&cart)

	if err != nil {
		return types.Cart{}, fmt.Errorf("%s: %w", op, err)
	}

	return cart, nil
}

// UpsertProduct removes product from cart if quantity eq zero.
// Adds product to cart if it is not found
// Update quantity if all is okay
func (s *MongoStore) UpsertProduct(
	ctx context.Context,
	userId string,
	productId string,
	quantity int,
) error {
	const op = "mongostorage.UpsertProduct"

	userOID, err := bson.ObjectIDFromHex(userId)
	if err != nil {
		return storage.ErrBadUserId
	}

	productOID, err := bson.ObjectIDFromHex(productId)
	if err != nil {
		return storage.ErrBadProductId
	}

	filter := bson.M{
		"user_id": userOID,
	}

	// 1. гарантируем, что корзина существует
	_, err = s.cartsCollection.UpdateOne(
		ctx,
		filter,
		bson.M{
			"$setOnInsert": bson.M{
				"user_id":  userOID,
				"products": []bson.M{},
			},
		},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// 2. если quantity == 0 → удаляем
	if quantity == 0 {
		_, err = s.cartsCollection.UpdateOne(
			ctx,
			filter,
			bson.M{
				"$pull": bson.M{
					"products": bson.M{
						"product_id": productOID,
					},
				},
			},
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	}

	// 3. пробуем обновить
	res, err := s.cartsCollection.UpdateOne(
		ctx,
		bson.M{
			"user_id":             userOID,
			"products.product_id": productOID,
		},
		bson.M{
			"$set": bson.M{
				"products.$.quantity": quantity,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// 4. если нет → добавляем
	if res.MatchedCount == 0 {
		_, err = s.cartsCollection.UpdateOne(
			ctx,
			filter,
			bson.M{
				"$push": bson.M{
					"products": bson.M{
						"product_id": productOID,
						"quantity":   quantity,
					},
				},
			},
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
