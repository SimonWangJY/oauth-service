package main

import (
	"log"
	"net/http"
	"oauth-service/pkg/consent"

	"github.com/gorilla/mux"
)

func main() {
	s := consent.NewServer()
	r := mux.NewRouter()

	r.HandleFunc("/consent", s.ConsentHandler)
	r.HandleFunc("/login", s.LoginHandler)
	r.HandleFunc("/callback", s.RedirectHandler)
	r.HandleFunc("/", s.HomePageHandler)

	log.Println("Listening on port 9091")
	http.ListenAndServe(":9091", r)
}
