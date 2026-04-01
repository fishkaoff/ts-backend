package types

import "go.mongodb.org/mongo-driver/v2/bson"

type UserRole string

const ADMIN UserRole = "admin"
const USER UserRole = "user"

type User struct {
	Id       bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Email    string        `json:"email" bson:"email"`
	PassHash []byte        `json:"-" bson:"pass_hash"`

	Name     string   `json:"name" bson:"name"`
	Lastname string   `json:"lastname" bson:"lastname"`
	Role     UserRole `json:"role" bson:"role"`
}

type RegisterDto struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Password string `json:"password"`
}

type LoginDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
