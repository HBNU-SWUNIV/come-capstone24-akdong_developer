package controllerdb

import (
	"GameServer/controller"
	"context"
	"fmt"

	"github.com/logrusorgru/aurora"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println(aurora.Sprintf(
			aurora.Red("MongoDB Connection Error : %s"), err))
		return
	}
	fmt.Println(aurora.Green("MongoDB Connection Success"))
	db := client.Database("GameServer")
	controller.DBClient = db
}
