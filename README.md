### Hydra Login and Consent Provider Implmented by Golang

This repo get the access token from hydra(oauth 2.0 server) through the browser and single api call
1. Browser: the most common way to get the access token. Mainly used in production environment and the application will redirect to certain front-end page to let user login/consent.
2. API Call: please be noted that this way only be used in development stage. We can easily get an access token withouth integrate with browser. This could be benefit for BE developer and tester.
