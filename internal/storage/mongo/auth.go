package mongostorage

import (
	"context"
	"fmt"

	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/fishkaoff/ts-backend/internal/storage"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func (s *MongoStore) SaveUser(ctx context.Context, user types.User) (types.User, error) {
	const op = "mongostorage.SaveUser"

	result, err := s.usersCollection.InsertOne(ctx, user)
	if err != nil {
		return types.User{}, fmt.Errorf("%s:%w", op, err)
	}

	user.Id = result.InsertedID.(bson.ObjectID)

	return user, nil
}

func (s *MongoStore) GetUserByEmail(ctx context.Context, email string) (types.User, error) {
	const op = "mongostorage.GetUserByEmail"
	filter := bson.D{
		{Key: "email", Value: email},
	}

	var user types.User

	err := s.usersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return types.User{}, storage.ErrUserNotFound
		}

		return types.User{}, fmt.Errorf("%s:%w", op, err)
	}

	return user, nil
}

func (s *MongoStore) GetUserById(ctx context.Context, id string) (types.User, error) {
	const op = "mongostorage.GetUserByEmail"

	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return types.User{}, fmt.Errorf("%s:%w", op, err)
	}

	filter := bson.D{
		{Key: "_id", Value: objId},
	}

	var user types.User

	err = s.usersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return types.User{}, storage.ErrUserNotFound
		}

		return types.User{}, fmt.Errorf("%s:%w", op, err)
	}

	return user, nil
}

func (s *MongoStore) DeleteUserByEmail(ctx context.Context, email string) error {
	panic("implement me")
}
