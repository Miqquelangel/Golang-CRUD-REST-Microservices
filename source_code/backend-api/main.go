package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"Signup"
	"Auth"
	"Login"
)

func main() {	

	l := log.New(os.Stdout, "\nproduct-api ", log.LstdFlags)
	sm := http.NewServeMux()
	signup := Signup.NewSignup(l)
	auth := Auth.NewAuth(l)
	login := Login.NewLogin(l)
	sm.Handle("/signup", signup) // Path to the login for the frontend // Only accepts POST and OPTIONS requests
	sm.Handle("/auth", auth) // Path to the login for the frontend // Only accepts POST and OPTIONS requests
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
