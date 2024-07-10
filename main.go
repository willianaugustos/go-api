package main

import (
	L "api-query/api/controllers"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Todo struct for MongoDB
type Todo struct {
	ID    string `json:"id" bson:"_id,omitempty"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var collection *mongo.Collection

func main() {

	logInitialize()

	// Set up MongoDB connection
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("testes").Collection("todos")

	// Define HTTP routes
	http.HandleFunc("/todos", getTodos)
	http.HandleFunc("/todos/add", addTodo)
	http.HandleFunc("/healthcheck", L.GetHealthCheck)

	// Start the server
	log.Println("Server listening on localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func logInitialize() {
	// Initialize Elasticsearch client
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	// Sample document to index
	doc := `{"name": "John Doe", "age": 30, "email": "john.doe@example.com"}`

	// Index a document into Elasticsearch
	res, err := es.Index(
		"my_index", // Index name
		esapi.ID(), // Document ID (auto-generated)
		//esapi strings.NewReader(doc),      // Document body
		esapi.WithDocumentType("_doc"), // Document type (for Elasticsearch < 7.x)
		esapi.WithContext(context.Background()),
	)
	if err != nil {
		log.Fatalf("Error indexing document: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error indexing document: %s", res.Status())
	} else {
		fmt.Println("Document indexed successfully!")
	}
}

// Handler to return all todos from MongoDB
func getTodos(w http.ResponseWriter, r *http.Request) {
	var todos []Todo

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.D{}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving todos", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var todo Todo
		err := cursor.Decode(&todo)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error decoding todos", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	if err := cursor.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Error with cursor while retrieving todos", http.StatusInternalServerError)
		return
	}

	// Respond with the list of todos as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// Handler to add a new todo to MongoDB
func addTodo(w http.ResponseWriter, r *http.Request) {
	var newTodo Todo

	// Decode JSON from request body
	err := json.NewDecoder(r.Body).Decode(&newTodo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, newTodo)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error inserting todo", http.StatusInternalServerError)
		return
	}

	log.Println("Inserted new todo with ID:", result.InsertedID)

	// Respond with the updated list of todos (including the new one)
	getTodos(w, r)
}
