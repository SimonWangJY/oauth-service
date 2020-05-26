package main

import (
	"log"
	"net/http"
	uiconsent "oauth-service/pkg/consent/ui_consent"

	"github.com/gorilla/mux"
)

func main() {
	s := uiconsent.NewServer()
	r := mux.NewRouter()

	// endpoint with UI page
	r.HandleFunc("/consent", s.ConsentHandler)
	r.HandleFunc("/login", s.LoginHandler)
	r.HandleFunc("/callback", s.RedirectHandler)
	r.HandleFunc("/", s.HomePageHandler)

	log.Println("Listening on port 9091")
	http.ListenAndServe(":9091", r)
}
