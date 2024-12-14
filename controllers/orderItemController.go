package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"restaurant_management/database"
	"restaurant_management/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderItemPack struct {
	Table_id    *string
	Order_items []models.OrderItem
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		result, err := orderItemCollection.Find(context.TODO(), bson.M{})
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items"})
			return
		}

		var allOrderItems []bson.M

		if err = result.All(ctx, &allOrderItems); err != nil {
			log.Fatal(err)
			return
		}

		c.JSON(http.StatusOK, allOrderItems)

	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

		orderId := c.Param("order_id")

		allOrderItems, err := ItemsByOrder(orderId)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order item"})
			return
		}

		c.JSON(http.StatusOK, allOrderItems)

	}
}

// func ItemsByOrder(id string) (OrderItems []models.OrderItem, err error) {
// 	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

// 	cursor, err := orderItemCollection.Find(context.TODO(), bson.M{"order_id": id})

// 	for cursor.Next(context.TODO()) {
// 		var result models.OrderItem
// 		if err1 := cursor.Decode(&result); err1 != nil {
// 			fmt.Println("Error decoding document:", err)
// 			return
// 		}
// 		OrderItems = append(OrderItems, result)
// 	}

// 	return OrderItems, err
// }

func ItemsByOrder(id string) (OrderItems []models.OrderItem, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel() // Ensures the context is canceled properly.

	matchStage := bson.D{{"$match", bson.D{{"order_id", id}}}}

	lookupStage := bson.D{{"$lookup", bson.D{
		{"from", "food"},
		{"localField", "food_id"},
		{"foreignField", "food_id"},
		{"as", "food"},
	}}}
	unwindStage := bson.D{{"$unwind", bson.D{
		{"path", "$food"},
		{"preserveNullAndEmptyArrays", true},
	}}}

	lookupOrderStage := bson.D{{"$lookup", bson.D{
		{"from", "order"},
		{"localField", "order_id"},
		{"foreignField", "order_id"},
		{"as", "order"},
	}}}
	unwindOrderStage := bson.D{{"$unwind", bson.D{
		{"path", "$order"},
		{"preserveNullAndEmptyArrays", true},
	}}}

	lookUpTableStage := bson.D{{"$lookup", bson.D{
		{"from", "table"},
		{"localField", "order.table_id"},
		{"foreignField", "table_id"},
		{"as", "table"},
	}}}
	unwindTableStage := bson.D{{"$unwind", bson.D{
		{"path", "$table"},
		{"preserveNullAndEmptyArrays", true},
	}}}

	projectStage := bson.D{{"$project", bson.D{
		{"_id", 0}, // Fixing "id" to "_id".
		{"amount", "$food.price"},
		{"total_count", 1},
		{"food_name", "$food.name"},
		{"food_image", "$food.food_image"},
		{"table_number", "$table.table_number"},
		{"table_id", "$table.table_id"},
		{"order_id", "$order.order_id"},
		{"price", "$food.price"},
		{"quantity", 1},
	}}}

	groupStage := bson.D{{"$group", bson.D{
		{"_id", bson.D{
			{"order_id", "$order_id"},
			{"table_id", "$table_id"},
			{"table_number", "$table_number"},
		}},
		{"payment_due", bson.D{{"$sum", "$amount"}}},
		{"total_count", bson.D{{"$sum", 1}}},
		{"order_items", bson.D{{"$push", "$$ROOT"}}}, // Fixed order_items aggregation.
	}}}

	projectStage2 := bson.D{{"$project", bson.D{
		{"_id", 0},
		{"payment_due", 1},
		{"total_count", 1},
		{"table_number", "$_id.table_number"},
		{"order_items", 1},
	}}}

	// Aggregate pipeline
	result, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupStage,
		unwindStage,
		lookupOrderStage,
		unwindOrderStage,
		lookUpTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		projectStage2,
	})
	if err != nil {
		return nil, err // Return the error properly instead of panicking.
	}

	// Decode the results
	if err = result.All(ctx, &OrderItems); err != nil {
		return nil, err // Return the error properly instead of panicking.
	}

	return OrderItems, nil
}




func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		orderItemId := c.Param("orderItem_id")

		var orderItem models.OrderItem

		err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order item"})
			return
		}

		c.JSON(http.StatusOK, orderItem)
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var orderItem models.OrderItem

		orderItemId := c.Param("orderItem_id")

		filter := bson.M{"order_item_id": orderItemId}

		var updateObj primitive.D

		if orderItem.Unit_Price != 0 {
			updateObj = append(updateObj, bson.E{"unit_price", orderItem.Unit_Price})
		}

		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{"quantity", orderItem.Quantity})
		}

		if orderItem.Food_id != nil {
			updateObj = append(updateObj, bson.E{"food_id", orderItem.Food_id})
		}

		orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", orderItem.Updated_at})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderItemCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := "Order Item not found"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		defer cancel()

		c.JSON(http.StatusOK, result)

	}
}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var orderItemPack OrderItemPack
		var order models.Order

		if err := c.BindJSON(&orderItemPack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "hb"})
			return
		}

		order.Order_Date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		orderItemsToBeInserted := []interface{}{}
		order.Table_id = orderItemPack.Table_id
		order_id := OrderItemOrderCreator(order)

		for _, orderItem := range orderItemPack.Order_items {
			orderItem.Order_id = order_id

			validationErr := validate.Struct(orderItem)

			if validationErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}

			orderItem.ID = primitive.NewObjectID()
			orderItem.Order_item_id = orderItem.ID.Hex()

			var foodItem models.Food

			foodCollection.FindOne(ctx, bson.M{"food_id": orderItem.Food_id}).Decode(&foodItem)

			// orderItem.Unit_Price = foodItem.Price
			// orderItem.Amount = (foodItem.Price * orderItem.Quantity)

			orderItem.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

			var num = toFixed(orderItem.Unit_Price, 2)
			orderItem.Unit_Price = num

			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
		}

		insertedOrderItems, err := orderItemCollection.InsertMany(ctx, orderItemsToBeInserted)

		if err != nil {
			log.Fatal("hi")
		}

		defer cancel()

		c.JSON(http.StatusOK, insertedOrderItems)

	}
}

func DeleteAllOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		// Specify the filter to identify the document(s) to delete

		// Delete the document matching the filter
		result, err := orderItemCollection.DeleteMany(ctx, bson.M{})
		if err != nil {
			fmt.Println("Error deleting document:", err)
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, result)

	}
}