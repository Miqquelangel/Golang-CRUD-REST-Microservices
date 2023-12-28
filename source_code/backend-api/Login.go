package Login

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

type Login struct {
	l *log.Logger
}

func NewLogin(l *log.Logger) *Login {
	return &Login{l}
}

var jwtKey = []byte("secret-key")

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var collection *mongo.Collection
var logCollection *mongo.Collection


func (login *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cookie, X-Token")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Expose-Headers", "*")

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
	logCollection = client.Database("logs").Collection("login")
	currentTime := time.Now()

		switch r.Method {
		case http.MethodPost:
			var loginData Item
			err := json.NewDecoder(r.Body).Decode(&loginData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
	
			filter := bson.D{{"username", loginData.Username}}
			var foundItem Item
			err = collection.FindOne(context.Background(), filter).Decode(&foundItem)
			if err != nil {
				logMessage := LogMessage{
					RequestURI: r.RequestURI,
					Status: http.StatusUnauthorized,
					IP:   r.RemoteAddr,
					Date: currentTime.Format("2006-01-02 15:04:05"),
					Header: fmt.Sprintf("%v", r.Header),
					Host: r.Host,
				}
				_, err = logCollection.InsertOne(context.Background(), logMessage)
				if err != nil {
					log.Fatal(err)
				}
				http.Error(w, "Invalid user", http.StatusBadRequest)
				return
			}
	
			if foundItem.Password != loginData.Password {
				logMessage := LogMessage{
					RequestURI: r.RequestURI,
					Status: http.StatusUnauthorized,
					IP:   r.RemoteAddr,
					Date: currentTime.Format("2006-01-02 15:04:05"),
					Header: fmt.Sprintf("%v", r.Header),
					Host: r.Host,
				}
				_, err = logCollection.InsertOne(context.Background(), logMessage)
				if err != nil {
					log.Fatal(err)
				}
				http.Error(w, "Invalid credentials", http.StatusBadRequest)
				return
			}
	
			expirationTime := time.Now().Add(time.Minute * 5)
			claims := &Claims{
				Username: foundItem.Username,
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
	
			http.SetCookie(w, &http.Cookie{
				Name: "token",
				Value: tokenString,
				Expires: expirationTime,
			})
	
			logMessage := LogMessage{
				RequestURI: r.RequestURI,
				Status: http.StatusOK,
				IP:   r.RemoteAddr,
				Date: currentTime.Format("2006-01-02 15:04:05"),
				Header: fmt.Sprintf("%v", r.Header),
				Host: r.Host,
			}
			_, err = logCollection.InsertOne(context.Background(), logMessage)
			if err != nil {
				log.Fatal(err)
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodOptions:
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.WriteHeader(http.StatusOK)
		}
	}
