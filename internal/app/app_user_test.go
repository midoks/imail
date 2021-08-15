package app

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

//map转字符串
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

//post access controller
func PostForm(uri string, param map[string]string, router *gin.Engine) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", uri+ParseToStr(param), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// go test -run TestIndex
func TestIndex(t *testing.T) {
	r := SetupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "hello world", w.Body.String())
}

// go test -run TestUserRegister
func TestUserRegister(t *testing.T) {

}

// go test -run TestUserLogin
func TestUserLogin(t *testing.T) {
	// var w *httptest.ResponseRecorder
	// assert := assert.New(t)
}
