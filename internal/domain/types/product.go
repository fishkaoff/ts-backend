package types

import (
	"strconv"
	"strings"

	"github.com/fishkaoff/ts-backend/internal/domain/lib/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductsFilter struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type Product struct {
	Id           bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Code         int           `json:"code" bson:"code"`
	PartNumber   string        `json:"part_number" bson:"part_number"`
	Name         string        `json:"name" bson:"name"`
	Manufacturer string        `json:"manufacturer" bson:"manufacturer"`
	Unit         ProductUnit   `json:"unit" bson:"unit"`
	Price        int64         `json:"price" bson:"price"`
	Balance      int64         `json:"balance" bson:"balance"`
	IsNew        bool          `json:"is_new" bson:"is_new"`
	ImageURL     string        `json:"image_url" bson:"image_url"`
	Active       bool          `json:"active" bson:"active"`
}

func ParsePrice(val string) int64 {
	val = strings.ReplaceAll(val, ",", ".")
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}
	return int64(f * 100)
}

func ParseIsNew(val string) bool {
	val = utils.NormalizeString(val)
	return val == "1"
}
