package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Menu struct {
	ID         primitive.ObjectID `bson:"_id"`
	Name       string             `json:"name" validate:"required"`
	Category   string             `json:"category" validate:"required"`
	Created_at time.Time         `json:"created_at"`
	Updated_at time.Time         `json:"updated_at"`
	Menu_id    string             `json:"menu_id"`
	Food_ids		[]string		`json:"food_ids"`
}
