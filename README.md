### Hydra Login and Consent Provider Implmented by Golang

This repo get the access token from hydra(oauth 2.0 server) through the browser and single api call
1. Browser: the most common way to get the access token. Mainly used in production environment and the application will redirect to certain front-end page to let user login/consent.
2. API Call: please be noted that this way only be used in development stage. We can easily get an access token withouth integrate with browser. This could be benefit for BE developer and tester.

### steps to start
1. run start.sh to setup the hydra
2. create hydra client copy the following code to the command line:
    hydra clients create \
    --endpoint http://127.0.0.1:4445 \
    --id my-app-client \
    --secret secret \
    --grant-types authorization_code,refresh_token \
    --response-types code,id_token \
    --scope openid,offline \
    --token-endpoint-auth-method client_secret_post \
    --callbacks http://127.0.0.1:9091/callback
3. for UI consent process, if you get 400 error response from chrome browser, try safari or other browser
4. if you want to test the authorisation flow from browser, execute 'go run main.go' in ui_consent folder. Then visit http://localhost:9091
5. if you want to run the authorisation flow in the BE and get token with provided login info, execute 'go run main.go' in consent folder. Then try to Post http request to http://127.0.0.1:9091/getToken with JSON body {"userName": "simon","password": "test"}. You could also run the server_test.go to test the result as well.
6. if you want to verify whether the access token is valid, post call http://localhost:4445/oauth2/introspect. add token key and put token value pair in x-www-form-urlencoded body type.