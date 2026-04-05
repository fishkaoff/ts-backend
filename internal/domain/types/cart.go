package types

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CartItem struct {
	ProductId bson.ObjectID `json:"product_id,omitempty" bson:"product_id"`
	Quantity  int           `json:"quantity,omitempty" bson:"quantity"`
}

type Cart struct {
	Id       bson.ObjectID `json:"id" bson:"_id"`
	UserId   bson.ObjectID `json:"user_id" bson:"user_id"`
	Products []CartItem    `json:"products" bson:"products,omitempty"`
}

type CartItemFull struct {
	Product  Product `json:"product"`
	Quantity int     `json:"quantity"`
}

type CartFull struct {
	Id       bson.ObjectID  `json:"id"`
	UserId   bson.ObjectID  `json:"user_id"`
	Products []CartItemFull `json:"products"`
}
