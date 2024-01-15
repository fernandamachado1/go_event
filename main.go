package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Event struct {
	mgm.DefaultModel `bson:",inline"`
	Name             string `json:"name"`
	Location         string `json:"location"`
	Description      string `json:"description"`
}

var client *mongo.Client

func createEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var event Event
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection := client.Database("Test").Collection("Event")

	_, err = collection.InsertOne(context.Background(), event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(event)
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := client.Database("Test").Collection("Event")

	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var events []Event
	for cursor.Next(context.Background()) {
		var event Event
		if err := cursor.Decode(&event); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		events = append(events, event)
	}

	json.NewEncoder(w).Encode(events)
}

func getEventCountByLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := client.Database("Test").Collection("Event")

	pipeline := bson.D{
		{"$group", bson.D{
			{"_id", "$location"},
			{"count", bson.D{{"$sum", 1}}},
		}},
	}

	cur, err := collection.Aggregate(context.Background(), mongo.Pipeline{pipeline})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())

	var results []bson.M
	if err := cur.All(context.Background(), &results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}

func main() {
	clientOptions := options.Client().ApplyURI("mongodb+srv://fernanda:12@events.wk9yycj.mongodb.net/")
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	r := mux.NewRouter()

	r.HandleFunc("/events", createEvent).Methods("POST")
	r.HandleFunc("/event-count-by-location", getEventCountByLocation).Methods("GET")
	r.HandleFunc("/events", getAllEvents).Methods("GET")

	port := "8080"
	addr := ":" + port
	fmt.Printf("Servidor ouvindo na porta %s...\n", port)
	http.ListenAndServe(addr, r)
}
