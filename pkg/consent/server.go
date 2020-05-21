package consent

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"net/url"

	hClient "github.com/ory/hydra-client-go/client"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
	"github.com/ory/x/randx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	host     = "localhost:4445"
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

var (
	scopes []string = []string{"openid", "offline"}
)

type templates struct {
	login    *template.Template
	consent  *template.Template
	home     *template.Template
	redirect *template.Template
}

func newTemplates() *templates {
	loginTpls := template.Must(template.New("").Parse(`<html>
	<body>
	<h1>Please sign in to proceed.</h1>
	<form action="/login?login_challenge={{.}}" method="POST">
	  <input name="username" type="text" />
	  <input name="password" type="password" /> 
	  <input type="submit" />
	  <table style="">
	  </table>
	</form>
	<p>To sign in, use the credentials "simon:test"</p>
	</body>
	</html>`))

	consentTpls := template.Must(template.New("").Parse(`<html>
	<body>
    <h1>The application wants access to:</h1>
    <form action="/consent?consent_challenge={{.consent_challenge}}" method="POST">
        <ul>
            {{range .requestedScopes}}
                <li><input type="checkbox" name="{{.}}">{{.}}</li>
            {{end}}
        </ul>
        <input type="submit">
    </form>
	</body>
	</html>`))

	homePageTpls := template.Must(template.New("").Parse(`<html>
	<body>
	<h1>Ory Hydra Login Login Test</h1>
	<p>To initiate the flow, click the "Authorize Application" button.</p>
	<p><a href="{{ . }}">Authorize application</a></p>
	</body>
	</html>`))

	redirectTpls := template.Must(template.New("").Parse(`<html>
	<html>
	<head></head>
	<body>
	<ul>
		<li>Access Token: <code>{{ .accessToken }}</code></li>
		<li>Refresh Token: <code>{{ .refreshToken }}</code></li>
		<li>Expires in: <code>{{ .expiry }}</code></li>
		<li>ID Token: <code>{{ .idtoken }}</code></li>
	</ul>
	</body>
	</html>`))

	return &templates{
		login:    loginTpls,
		consent:  consentTpls,
		home:     homePageTpls,
		redirect: redirectTpls,
	}
}

type Server struct {
	hydra     *hClient.OryHydra
	templates *templates
	// oauth2Client *models.OAuth2Client
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
	// check client is exist. if not register a oauth 2 client to the authorization server
	// getOAuth2ClientResp, err := hydra.Admin.GetOAuth2Client(&admin.GetOAuth2ClientParams{
	// 	ID:      "abc",
	// 	Context: context.Background(),
	// })
	// if err != nil {
	// 	return Server{}, err
	// }

	// log.Println(getOAuth2ClientResp)

	return &Server{
		hydra:        hydra,
		templates:    newTemplates(),
		oauth2Config: oConfig,
	}
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LoginHandler...")
	challengeID := r.URL.Query().Get("login_challenge")
	if r.Method == http.MethodGet {
		if err := s.templates.login.Execute(w, challengeID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userName := r.Form.Get("username")
	password := r.Form.Get("password")

	if userName != "simon" || password != "test" {
		http.Error(w, "provided credentials are wrong, try simon:test", http.StatusBadRequest)
		return
	}

	loginReq, err := s.hydra.Admin.GetLoginRequest(&admin.GetLoginRequestParams{
		LoginChallenge: challengeID,
		Context:        context.Background(),
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(loginReq)

	acceptLoginRsp, err := s.hydra.Admin.AcceptLoginRequest(&admin.AcceptLoginRequestParams{
		LoginChallenge: loginReq.Payload.Challenge,
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

func (s *Server) ConsentHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ConsentHandler...")
	challengeID := r.URL.Query().Get("consent_challenge")
	if r.Method == http.MethodGet {
		consentReq, err := s.hydra.Admin.GetConsentRequest(&admin.GetConsentRequestParams{
			ConsentChallenge: challengeID,
			Context:          context.Background(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("consent request", consentReq)

		tplData := map[string]interface{}{
			"consent_challenge": challengeID,
			"requestedScopes":   consentReq.Payload.RequestedScope,
		}

		if err := s.templates.consent.Execute(w, tplData); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

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
	log.Println("accept consent response", acceptConsentResp)

	http.Redirect(w, r, acceptConsentResp.Payload.RedirectTo, http.StatusFound)
}

func (s *Server) HomePageHandler(w http.ResponseWriter, r *http.Request) {
	state, err := randx.RuneSequence(24, randx.AlphaLower)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authURL := s.oauth2Config.AuthCodeURL(string(state))
	log.Println(authURL)
	if err := s.templates.home.Execute(w, authURL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *Server) RedirectHandler(w http.ResponseWriter, r *http.Request) {
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

	data := map[string]interface{}{
		"accessToken":  token.AccessToken,
		"refreshToken": token.RefreshToken,
		"expiry":       token.Expiry,
		"idtoken":      token.Extra("id_token"),
	}
	if err := s.templates.redirect.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
