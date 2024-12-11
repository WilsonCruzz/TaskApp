package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
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
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

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

func createTodo(c *fiber.Ctx) error {
	// Create a new Todo object to store the parsed data
	todo := new(Todo)

	// Parse the request body into the Todo object
	if err := c.BodyParser(todo); err != nil {
		// If there is an error while parsing the body, return the error
		return err
	}

	// Check if the Body field of the Todo is empty
	if todo.Body == "" {
		// If the Body is empty, return a 400 Bad Request with an error message
		return c.Status(400).JSON(fiber.Map{"error": errors.New("Body is required")})
	}

	// Insert the new Todo into the MongoDB collection
	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		// If there is an error during the insertion, return the error
		return err
	}

	// Set the ID field of the Todo with the inserted document's ID
	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	// Return a 201 Created status with the created Todo in the response body
	return c.Status(201).JSON(todo)
}

func updateTodo(c *fiber.Ctx) error {
	// Retrieve the "id" parameter from the URL
	id := c.Params("id")

	// Convert the "id" string to a MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// If the ObjectID conversion fails, return a 400 Bad Request error with a message
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	// Define the filter to find the Todo item by its ObjectID
	filter := bson.M{"_id": objectID}

	// Define the update operation to set the "completed" field to true
	update := bson.M{"$set": bson.M{"completed": true}}

	// Execute the update operation in the MongoDB collection
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		// If there's an error during the update, return the error
		return err
	}

	// Return a 200 OK status with a success message
	return c.Status(200).JSON(fiber.Map{"message": "Todo updated successfully"})
}

func deleteTodo(c *fiber.Ctx) error {
	// Retrieve the "id" parameter from the URL
	id := c.Params("id")

	// Convert the "id" string to a MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// If the ObjectID conversion fails, return a 400 Bad Request error with a message
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	// Define the filter to find the Todo item by its ObjectID
	filter := bson.M{"_id": objectID}

	// Execute the delete operation to remove the Todo item from the collection
	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		// If there's an error during the deletion, return the error
		return err
	}

	// Return a 200 OK status with a success message indicating the Todo was deleted
	return c.Status(200).JSON(fiber.Map{"message": "Todo deleted successfully"})
}
