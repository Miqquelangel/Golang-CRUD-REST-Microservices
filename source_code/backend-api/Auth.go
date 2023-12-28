package Auth

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/dgrijalva/jwt-go"
)

var logCollection *mongo.Collection
var jwtKey = []byte("secret-key")

type LogMessage struct {
	RequestURI string `json:"requestURI"`
	Status int `json:"status"`
	IP     string `json:"ip"`
	Date   string `json:"date"`
	Header string `json:"header"`
	Host   string `json:"host"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type Auth struct {
	l *log.Logger
}

func NewAuth(l *log.Logger) *Auth {
	return &Auth{l}
}

func (auth *Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "http://192.168.49.2:30010")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cookie")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Expose-Headers", "*")

	switch r.Method {
	
	case http.MethodPost:
	clientOptions := options.Client().ApplyURI("mongodb://192.168.49.2:30001")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

		logCollection = client.Database("logs").Collection("auth")
		cookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tokenStr := cookie.Value

	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
	func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	currentTime := time.Now()
			logMessage := LogMessage{
				RequestURI: r.RequestURI,
				Status: http.StatusOK,
				IP:   r.RemoteAddr,
				Date: currentTime.Format("2006-01-02 15:04:05"),
				Header: fmt.Sprintf("%v", r.Header),
				Host: r.Host,
			}
			a, err := logCollection.InsertOne(context.Background(), logMessage)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Inserted a single document: ", a.InsertedID)

	w.Write([]byte(fmt.Sprintf("\nHello! User (%s) has been authenticated\n", claims.Username)))
	}

}
