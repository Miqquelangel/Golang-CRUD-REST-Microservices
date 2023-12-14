package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"os/signal"
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


type Item struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type Id struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Id string `json:"_id"` 
}

type LogMessage struct {
	RequestURI string `json:"requestURI"`
	Status   int  `json:"status"`
	IP       string `json:"ip"`
	Date     string `json:"date"`
	Header   string `json:"header"`
	Host     string `json:"host"`
	ResponseHeader string `json:"responseHeader"`
}

type Logging struct {
	ll *log.Logger
}

func newLogging(ll *log.Logger) *Logging {
	return &Logging{ll}
}

type LogProducts struct {
	l *log.Logger
}

func NewProducts(l *log.Logger) *LogProducts {
	return &LogProducts{l}
}

type Login struct {
	l *log.Logger
}

func NewLogin(l *log.Logger) *Login {
	return &Login{l}
}

var collection *mongo.Collection
var logCollection *mongo.Collection

func main() {

	// Setup where we are going to connect to the DB.
	clientOptions := options.Client().ApplyURI("mongodb://dev-mongo-service.mongo.svc.cluster.local:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("login").Collection("users")
	logCollection = client.Database("Logs").Collection("Api")

	l := log.New(os.Stdout, "\nproduct-api ", log.LstdFlags)
	sm := http.NewServeMux()
	hh := NewProducts(l)
	hhh := newLogging(l)
	login := NewLogin(l)
	sm.Handle("/logs", hhh) // Path to the logs of succesful / unsuccesful login attempts
	sm.Handle("/users", hh) // Path to interact to the DB via curl commands
	sm.Handle("/login", login) // Path to the login for the frontend // Only accepts POST and OPTIONS requests
	s := &http.Server{
		Addr:         "0.0.0.0:80",
		Handler:      sm,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			l.Fatal(err)
		}
	}()
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)
	sig := <-sigChan
	fmt.Println("graceful shutdown", sig)
	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)
}

func (h *LogProducts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	switch r.Method {
	case http.MethodGet:
		// Handle GET request
		var item Item
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		filter := bson.D{{"username", item.Username}, {"password", item.Password}}
		cursor, err := collection.Find(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if cursor.Next(context.Background()) {
			// User exists, login successful
			w.WriteHeader(http.StatusOK)
			currentTime := time.Now()
			logMessage := LogMessage{
				RequestURI: r.RequestURI,
				Status:   http.StatusOK,
				IP:       r.RemoteAddr,
				Date:     currentTime.Format("2006-01-02 15:04:05"),
				Header:   fmt.Sprintf("%v", r.Header),
				Host:     r.Host,
				ResponseHeader: fmt.Sprintf("%v", w.Header()),
			}
			insertResult, err := logCollection.InsertOne(context.Background(), logMessage)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Inserted a single document: ", insertResult.InsertedID)
			logMessag := fmt.Sprintf(" MongoID: %s Succesful request to path %s HTTP Status: %d IP: %s Date: %s Header: %s Host: %s ResponseHeader: %s\n", insertResult.InsertedID, r.RequestURI, http.StatusOK, r.RemoteAddr, currentTime.Format("2006-01-02 15:04:05"), r.Header, r.Host, w.Header)
			h.l.Println(logMessag)
		} else {
			// User does not exist, login failed
			currentTime := time.Now()
			logMessage := LogMessage{
				RequestURI: r.RequestURI,
				Status:   http.StatusUnauthorized,
				IP:       r.RemoteAddr,
				Date:     currentTime.Format("2006-01-02 15:04:05"),
				Header:   fmt.Sprintf("%v", r.Header),
				Host:     r.Host,
				ResponseHeader: fmt.Sprintf("%v", w.Header()),
			}
			insertResult, err := logCollection.InsertOne(context.Background(), logMessage)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Inserted a single document: ", insertResult.InsertedID)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			fmt.Printf("cagaste\n")
		}
	case http.MethodOptions:
		// Handle OPTIONS request
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.WriteHeader(http.StatusOK)
	 
	case http.MethodPut:
		// Handle PUT requests
		var userPut Id
		err := json.NewDecoder(r.Body).Decode(&userPut)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := primitive.ObjectIDFromHex(userPut.Id)
		fmt.Println(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		filter := bson.D{{"_id", id}}
		update := bson.D{
			{"$set", bson.D{
				{"username", userPut.Username},
				{"password", userPut.Password},
			}},
		}
		updateResult, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if updateResult.ModifiedCount == 0 {
			http.Error(w, "No document was updated", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		// Handle DELETE requests
		var userDelete Id
		err := json.NewDecoder(r.Body).Decode(&userDelete)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := primitive.ObjectIDFromHex(userDelete.Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		filter := bson.D{{"_id", id}}
		deleteResult, err := collection.DeleteOne(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if deleteResult.DeletedCount == 0 {
			http.Error(w, "No document was deleted", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodPost:
		// Handle POST requests
		var userCreate Item
		err := json.NewDecoder(r.Body).Decode(&userCreate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		insertResult, err := collection.InsertOne(context.Background(), userCreate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(insertResult.InsertedID)
		

	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

	
}

func (hh *Logging) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Handle GET request
		items := []LogMessage{}
		cursor, err := logCollection.Find(context.Background(), bson.D{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for cursor.Next(context.Background()) {
			var item LogMessage
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}
		json.NewEncoder(w).Encode(items)

	case http.MethodOptions:
		// Handle OPTIONS request
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func (login *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	switch r.Method {
	case http.MethodPost:
		// Handle POST request
		var item Item
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		filter := bson.D{{"username", item.Username}, {"password", item.Password}}
		cursor, err := collection.Find(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if cursor.Next(context.Background()) {
			// User exists, login successful
			w.WriteHeader(http.StatusOK)
			currentTime := time.Now()
			logMessage := LogMessage{
				RequestURI: r.RequestURI,
				Status:   http.StatusOK,
				IP:       r.RemoteAddr,
				Date:     currentTime.Format("2006-01-02 15:04:05"),
				Header:   fmt.Sprintf("%v", r.Header),
				Host:     r.Host,
				ResponseHeader: fmt.Sprintf("%v", w.Header()),
			}
			insertResult, err := logCollection.InsertOne(context.Background(), logMessage)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Inserted a single document: ", insertResult.InsertedID)
			logMessag := fmt.Sprintf(" MongoID: %s Succesful request to path %s HTTP Status: %d IP: %s Date: %s Header: %s Host: %s ResponseHeader: %s\n", insertResult.InsertedID, r.RequestURI, http.StatusOK, r.RemoteAddr, currentTime.Format("2006-01-02 15:04:05"), r.Header, r.Host, w.Header)
			login.l.Println(logMessag)

		} else {
			// User does not exist, login failed
			currentTime := time.Now()
			logMessage := LogMessage{
				RequestURI: r.RequestURI,
				Status:   http.StatusUnauthorized,
				IP:       r.RemoteAddr,
				Date:     currentTime.Format("2006-01-02 15:04:05"),
				Header:   fmt.Sprintf("%v", r.Header),
				Host:     r.Host,
				ResponseHeader: fmt.Sprintf("%v", w.Header()),
			}
			insertResult, err := logCollection.InsertOne(context.Background(), logMessage)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Inserted a single document: ", insertResult.InsertedID)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			fmt.Printf("cagaste\n")
		}
	case http.MethodOptions:
		// Handle OPTIONS request
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.WriteHeader(http.StatusOK)
	}
}
