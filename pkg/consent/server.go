// Package consent used to get the oauth token  from backend
package consent

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	hClient "github.com/ory/hydra-client-go/client"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
	"github.com/ory/x/randx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	host     = "127.0.0.1:4445"
	basePath = "/"
	scheme   = "http"
)

const (
	clientID            = "my-app-client"
	clientSecret        = "secret"
	oAuthPublicTokenURL = "http://127.0.0.1:4444/oauth2/token"
	oAuthPublicAuthURL  = "http://127.0.0.1:4444/oauth2/auth"
	redirectURL         = "http://127.0.0.1:9091/callback"
)

const (
	subjectName = "Username"
)

var (
	scopes []string = []string{"openid", "offline"}
)

// LoginRequest user login request
type LoginRequest struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

// LoginResponse user login response
type LoginResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	Expiry       time.Time `json:"expiry"`
	IDToken      string    `json:"idToken"`
}

type Server struct {
	hydra        *hClient.OryHydra
	oauth2Config *oauth2.Config
}

// NewServer get new login/consent server
func NewServer() *Server {
	cfg := &hClient.TransportConfig{
		Host:     host,
		Schemes:  []string{scheme},
		BasePath: basePath,
	}

	oConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oAuthPublicAuthURL,
			TokenURL: oAuthPublicTokenURL,
		},
		RedirectURL: redirectURL,
	}

	hydra := hClient.NewHTTPClientWithConfig(nil, cfg)

	return &Server{
		hydra:        hydra,
		oauth2Config: oConfig,
	}
}

func (s *Server) HydraLoginProviderHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LoginHandler...")

	userName := r.Header.Get(subjectName)
	if len(userName) == 0 {
		http.Error(w, "user name cannot be empty", http.StatusBadRequest)
		return
	}

	challengeID := r.URL.Query().Get("login_challenge")
	if len(challengeID) == 0 {
		http.Error(w, "missing challenge ID", http.StatusBadRequest)
		return
	}

	acceptLoginRsp, err := s.hydra.Admin.AcceptLoginRequest(&admin.AcceptLoginRequestParams{
		LoginChallenge: challengeID,
		Body: &models.AcceptLoginRequest{
			Subject: &userName,
		},
		Context: context.Background(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, acceptLoginRsp.Payload.RedirectTo, http.StatusFound)
}

func (s *Server) HydraConsentProviderHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ConsentProviderHandler...")

	challengeID := r.URL.Query().Get("consent_challenge")

	acceptConsentResp, err := s.hydra.Admin.AcceptConsentRequest(&admin.AcceptConsentRequestParams{
		ConsentChallenge: challengeID,
		Body: &models.AcceptConsentRequest{
			GrantScope: []string{"offline_access", "offline", "openid"},
		},
		Context: context.Background(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, acceptConsentResp.Payload.RedirectTo, http.StatusFound)
}

func (s *Server) HydraRedirectHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HydraRedirectHandler...")

	code := r.URL.Query().Get("code")
	oauthConf := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     oAuthPublicTokenURL,
		Scopes:       scopes,
		EndpointParams: url.Values{
			"grant_type":   {"authorization_code"},
			"redirect_uri": {redirectURL},
			"client_id":    {clientID},
			"code":         {code}},
	}

	token, err := oauthConf.Token(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	loginResp := LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		IDToken:      token.Extra("id_token").(string),
	}

	loginRespBytes, err := json.Marshal(loginResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(loginRespBytes)
}

func (s *Server) GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("GetTokenHandler...")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var loginReq LoginRequest
	if err := json.Unmarshal(bodyBytes, &loginReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// the reason to include user name into header is because we need to input user name as
	// subject in login provider.
	userName := r.Header.Get(subjectName)
	if len(userName) == 0 || userName != loginReq.UserName {
		http.Error(w, "user name must be included in header", http.StatusBadRequest)
		return
	}

	// login credential validation
	if loginReq.UserName != "simon" || loginReq.Password != "test" {
		http.Error(w, "user login info incorrect", http.StatusBadRequest)
		return
	}

	state, err := randx.RuneSequence(24, randx.AlphaLower)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authURL := s.oauth2Config.AuthCodeURL(string(state))
	http.Redirect(w, r, authURL, http.StatusFound)
	return
}
