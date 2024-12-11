package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        int    `json:"_id" bson:"_id"`
	Completed bool   `json:"completed"`
	Body      string `json:"body"`
}

var collection *mongo.Collection

func main() {
	fmt.Println("Hello, World!")

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")
	clientOption := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOption)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection = (client.Database("golang_db").Collection("todos"))

	app := fiber.New()

	app.Get("/api/todos", getTodos)
	// app.Post("/api/todos", createTodo)
	// app.Patch("/api/todos", updateTodo)
	// app.Delete("/api/todos", deleteTodo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getTodos(c *fiber.Ctx) error {
	// Declare a slice to hold the list of todos
	var todos []Todo

	// Perform a query to find all documents in the collection (no filter)
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		// If an error occurs during the query, return the error
		return err
	}

	// Ensure the cursor is closed when the function finishes
	defer cursor.Close(context.Background())

	// Iterate over the documents returned by the query
	for cursor.Next(context.Background()) {
		// Declare a variable to hold each todo item
		var todo Todo

		// Decode the current document into the todo variable
		if err := cursor.Decode(&todo); err != nil {
			// If decoding fails, return the error
			return err
		}

		// Append the decoded todo to the todos slice
		todos = append(todos, todo)
	}

	// Return the todos list as a JSON response
	return c.JSON(todos)
}

// func createTodo(c *fiber.Ctx) error {}
// func updateTodo(c *fiber.Ctx) error {}
// func deleteTodo(c *fiber.Ctx) error {}
