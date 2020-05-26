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

	// BE endpoint
	r.HandleFunc("/consent", s.HydraConsentProviderHandler)
	r.HandleFunc("/login", s.HydraLoginProviderHandler).Methods(http.MethodGet)
	r.HandleFunc("/callback", s.HydraRedirectHandler)
	r.HandleFunc("/getToken", s.GetTokenHandler).Methods(http.MethodPost)

	log.Println("Listening on port 9091")
	http.ListenAndServe(":9091", r)
}
