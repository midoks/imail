package app

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/midoks/imail/internal/libs"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

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

// go test -run TestIndex
func TestIndex(t *testing.T) {
	r := SetupRouter()
	w := Get("/", r)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "hello world", w.Body.String())
}

// go test -run TestUserRegister
func TestUserRegister(t *testing.T) {

}

// go test -run TestUserLogin
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
	postBody.Add("password", libs.Md5str(libs.Md5str(password)+wcode["rand"]))

	// fmt.Println("in", password, wcode["rand"])

	w = PostForm("/v1/login", postBody, r)

	var result map[string]string
	_ = json.Unmarshal([]byte(w.Body.String()), &result)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 32, len(result["token"]))
}
