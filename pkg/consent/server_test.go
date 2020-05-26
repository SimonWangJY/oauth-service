package consent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type jar struct {
	lk      sync.Mutex
	cookies map[string][]*http.Cookie
}

func newJar() *jar {
	jar := new(jar)
	jar.cookies = make(map[string][]*http.Cookie)
	return jar
}

// SetCookies handles the receipt of the cookies in a reply for the
// given URL.  It may or may not choose to save the cookies, depending
// on the jar's policy and implementation.
func (jar *jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lk.Lock()
	jar.cookies[u.Host] = cookies
	jar.lk.Unlock()
}

// Cookies returns the cookies to send in a request for the given URL.
// It is up to the implementation to honor the standard cookie use
// restrictions such as in RFC 6265.
func (jar *jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
}

func Test_Sever(t *testing.T) {
	loginReq := LoginRequest{
		UserName: "simon",
		Password: "test",
	}

	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:9091/getToken", getReader(loginReq))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("UserName", loginReq.UserName)

	jar := newJar()
	client := http.Client{Jar: jar}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var loginResp LoginResponse
	json.Unmarshal(body, &loginResp)

	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, loginResp.Expiry)
	assert.NotEmpty(t, loginResp.IDToken)
	assert.NotEmpty(t, loginResp.RefreshToken)
}

func getReader(obj interface{}) io.Reader {
	reader := new(bytes.Buffer)
	json.NewEncoder(reader).Encode(obj)
	fmt.Println(reader.String())
	return reader
}
