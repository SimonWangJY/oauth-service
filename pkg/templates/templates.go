package templates

import "text/template"

type Templates struct {
	Login    *template.Template
	Consent  *template.Template
	Home     *template.Template
	Redirect *template.Template
}

func NewTemplates() *Templates {
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
	<h1>Ory Hydra Login Test</h1>
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

	return &Templates{
		Login:    loginTpls,
		Consent:  consentTpls,
		Home:     homePageTpls,
		Redirect: redirectTpls,
	}
}
