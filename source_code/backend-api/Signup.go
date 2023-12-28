package Signup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"go.mongodb.org/mongo-driver/bson"
	"time"
	"log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/go-playground/validator/v10"
	"github.com/dgrijalva/jwt-go"
)

type Item struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LogMessage struct {
	RequestURI string `json:"requestURI"`
	Status int `json:"status"`
	IP     string `json:"ip"`
	Date   string `json:"date"`
	Header string `json:"header"`
	Host   string `json:"host"`
}

type Signup struct {
	l *log.Logger
}

func NewSignup(l *log.Logger) *Signup {
	return &Signup{l}
}

var jwtKey = []byte("secret-key")

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var collection *mongo.Collection
var logCollection *mongo.Collection

func (signup *Signup) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cookie, X-Token")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Expose-Headers", "*")

	start := time.Now() // Record the start time
	clientOptions := options.Client().ApplyURI("mongodb://192.168.49.2:30001")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("users").Collection("data")
	logCollection = client.Database("logs").Collection("signup")

	switch r.Method {
	
	case http.MethodPost:
		var item Item
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	
		validate := validator.New()
		err = validate.Struct(item)
		if err != nil {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			for _, err := range err.(validator.ValidationErrors) {
				if err.Field() == "Username" && err.Tag() == "required" {
					http.Error(w, "username empty", http.StatusBadRequest)
					return
				}
				if err.Field() == "Password" && err.Tag() == "required" {
					http.Error(w, "pass empty", http.StatusBadRequest)
					return
				}
			}
		}
	
		filter := bson.D{{"username", item.Username}}
		cursor, err := collection.Find(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		if cursor.Next(context.Background()) {
			http.Error(w, "Invalid user", http.StatusBadRequest)
			return
		} else {
			expirationTime := time.Now().Add(time.Minute * 5)
			claims := &Claims{
				Username: item.Username,
				StandardClaims: jwt.StandardClaims {
					ExpiresAt: expirationTime.Unix(),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString(jwtKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.SetCookie (w,
				&http.Cookie{
					Name: "token",
					Value: tokenString,
					Expires: expirationTime,
				})
			w.WriteHeader(http.StatusCreated)
			insertResult, err := collection.InsertOne(context.Background(), item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Println("\nInserted a single document: ", insertResult.InsertedID)
			currentTime := time.Now()
			logMessage := LogMessage{
				RequestURI: r.RequestURI,
				Status: http.StatusOK,
				IP:   r.RemoteAddr,
				Date: currentTime.Format("2006-01-02 15:04:05"),
				Header: fmt.Sprintf("%v", r.Header),
				Host: r.Host,
			}
			insertResult, err = logCollection.InsertOne(context.Background(), logMessage)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Inserted a single document: ", insertResult.InsertedID)
			end := time.Now() // Record the end time
			duration := end.Sub(start) // Calculate the duration
			fmt.Printf("The database operation took %v\n", duration)
		}
	
		case http.MethodOptions:
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.WriteHeader(http.StatusOK)
		}
		
	}

