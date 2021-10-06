package app

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/tools"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// go test -v ./internal/app
func init() {
	os.MkdirAll("./data", 0777)
	os.MkdirAll("./logs", 0777)

	err := conf.Load("../../conf/app.defined.conf")
	if err != nil {
		fmt.Println("init config fail:", err.Error())
	}

	log.Init()
	db.Init()
	gin.SetMode(gin.ReleaseMode)
	go Start(1180)

	time.Sleep(1 * time.Second)
}

// go test -v ./internal/app

var token string

//map to string
func ParseToStr(mp map[string]string) string {
	values := ""
	for key, val := range mp {
		values += "&" + key + "=" + val
	}
	temp := values[1:]
	values = "?" + temp
	return values
}

//get access controller
func Get(uri string, router *gin.Engine) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", uri, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

//post query access controller
//exp: PostFormQuery("/v1/login", map[string]string{"name": "admin", "password": "admin"}, r)
func PostFormQuery(uri string, param map[string]string, router *gin.Engine) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", uri+ParseToStr(param), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

//post access controller
func PostForm(uri string, param url.Values, router *gin.Engine) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", uri, strings.NewReader(param.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func initToken() string {
	r := SetupRouter()
	if token != "" {
		return token
	}

	user := "admin"
	password := "admin"

	w := Get("/v1/get_code?name="+user, r)
	var wcode map[string]string
	_ = json.Unmarshal([]byte(w.Body.String()), &wcode)

	postBody := make(url.Values)
	postBody.Add("name", user)
	postBody.Add("token", wcode["token"])
	postBody.Add("password", tools.Md5(tools.Md5(password)+wcode["rand"]))

	w = PostForm("/v1/login", postBody, r)

	var result map[string]string
	_ = json.Unmarshal([]byte(w.Body.String()), &result)

	token = result["token"]
	return token
}

// go test -run TestIndex
func TestIndex(t *testing.T) {
	r := SetupRouter()
	w := Get("/", r)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "hello world", w.Body.String())
}

// go test -run TestUserRegister
func TestUserRegister(t *testing.T) {
	assert.Equal(t, 200, 200)
}

/// go test -run TestUserLogin
func TestUserLogin(t *testing.T) {
	r := SetupRouter()

	user := "admin"
	password := "admin"

	w := Get("/v1/get_code?name="+user, r)
	var wcode map[string]string
	_ = json.Unmarshal([]byte(w.Body.String()), &wcode)

	// fmt.Println(wcode["token"])
	// fmt.Println(wcode["rand"])

	postBody := make(url.Values)
	postBody.Add("name", user)
	postBody.Add("token", wcode["token"])
	postBody.Add("password", tools.Md5(tools.Md5(password)+wcode["rand"]))

	// fmt.Println("in", password, wcode["rand"])

	w = PostForm("/v1/login", postBody, r)

	var result map[string]string
	_ = json.Unmarshal([]byte(w.Body.String()), &result)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 32, len(result["token"]))
}

// go test -run TestToken
func TestToken(t *testing.T) {
	token := initToken()
	assert.Equal(t, 32, len(token))
}

//go test -bench=. -benchmem ./...
//go test -bench=. -benchmem ./internal/app
func BenchmarkGetToken(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := initToken()
		assert.Equal(b, 32, len(token))
	}
	b.StopTimer()
}
