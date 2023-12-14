package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"Login"
)

func main() {	

	l := log.New(os.Stdout, "\nproduct-api ", log.LstdFlags)
	sm := http.NewServeMux()
	login := Login.NewLogin(l)
	status := Login.NewHome(l)
	sm.Handle("/login", login) // Path to the login for the frontend // Only accepts POST and OPTIONS requests
	sm.Handle("/status", status) // Path to the login for the frontend // Only accepts POST and OPTIONS request
	s := &http.Server{
		Addr:         "10.211.55.26:80",
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


