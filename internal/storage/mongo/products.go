package mongostorage

import (
	"context"
	"fmt"

	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/storage"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *MongoStore) GetProducts(ctx context.Context, filter types.ProductsFilter) ([]types.Product, error) {
	const op = "mongostorage.GetProducts"

	options := options.Find().SetSkip(int64(filter.Offset)).SetLimit(int64(filter.Limit))
	mongoFilter := bson.D{{Key: "active", Value: true}}

	cursor, err := s.productsCollection.Find(ctx, mongoFilter, options)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []types.Product{}, nil
		}
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	var result []types.Product
	if err := cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return result, nil
}

func (s *MongoStore) GetProductById(ctx context.Context, id string) (types.Product, error) {
	const op = "mongostorage.GetProductById"

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return types.Product{}, fmt.Errorf("%s: invalid id: %w", op, err)
	}

	filter := bson.D{{Key: "_id", Value: objID}}

	var product types.Product
	err = s.productsCollection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return types.Product{}, storage.ErrDocumentNotFound
		}

		return types.Product{}, fmt.Errorf("%s:%w", op, err)
	}

	return product, nil
}

func (s *MongoStore) GetProductsByIds(ctx context.Context, ids []bson.ObjectID) ([]types.Product, error) {
	const op = "mongostorage.GetProductsByIds"

	if len(ids) == 0 {
		return []types.Product{}, nil
	}

	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}

	cursor, err := s.productsCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var products []types.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return products, nil
}

func (s *MongoStore) GetProductsByPartNumbers(ctx context.Context, partNumbers []string) ([]types.Product, error) {
	const op = "mongostorage.GetProductsByIds"

	if len(partNumbers) == 0 {
		return []types.Product{}, nil
	}

	filter := bson.M{
		"part_number": bson.M{
			"$in": partNumbers,
		},
	}

	cursor, err := s.productsCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var products []types.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return products, nil
}

func (s *MongoStore) UpdateProducts(ctx context.Context, products []types.Product) error {
	const op = "mongostorage.UpdateProducts"

	// collect active part numbers
	partNumbers := make([]string, 0, len(products))
	for _, p := range products {
		partNumbers = append(partNumbers, p.PartNumber)
	}

	// update or insert products
	if err := s.batchProductsUpdate(ctx, products); err != nil {
		return fmt.Errorf("%s: batchProductsUpdate failed: %w", op, err)
	}

	// mark inactive products
	if err := s.markInactiveProducts(ctx, partNumbers); err != nil {
		return fmt.Errorf("%s: markInactiveProducts failed: %w", op, err)
	}

	return nil
}

func (s *MongoStore) batchProductsUpdate(ctx context.Context, products []types.Product) error {
	const op = "mongostore.batchProducsUpdate"
	const batchSize = 1000

	for i := 0; i < len(products); i += batchSize {
		end := i + batchSize
		if end > len(products) {
			end = len(products)
		}

		batch := products[i:end]
		models := make([]mongo.WriteModel, 0, len(batch))

		for _, p := range batch {
			filter := bson.M{"part_number": p.PartNumber}

			// update all fields or insert new product
			update := bson.M{
				"$set": bson.M{
					"code":         p.Code,
					"name":         p.Name,
					"manufacturer": p.Manufacturer,
					"unit":         p.Unit,
					"price":        p.Price,
					"balance":      p.Balance,
					"is_new":       p.IsNew,
					"image_url":    p.ImageURL,
					"active":       true,
				},
			}

			model := mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(true)

			models = append(models, model)
		}

		opts := options.BulkWrite().SetOrdered(false)
		_, err := s.productsCollection.BulkWrite(ctx, models, opts)
		if err != nil {
			return fmt.Errorf("%s: bulk upsert failed: %w", op, err)
		}
	}

	return nil
}

func (s *MongoStore) markInactiveProducts(ctx context.Context, newPartNumbers []string) error {
	const op = "mongostorage.markIncativeProducts"

	_, err := s.productsCollection.UpdateMany(ctx,
		bson.M{"part_number": bson.M{"$nin": newPartNumbers}},
		bson.M{"$set": bson.M{"active": false}},
	)

	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}
