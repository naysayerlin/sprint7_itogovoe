package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"
	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: min(len(cafeList[city]), 100)},
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest(
			"GET",
			"/cafe?city="+city+"&count="+strconv.Itoa(v.count),
			nil,
		)
		handler.ServeHTTP(response, req)
		assert.Equal(t, http.StatusOK, response.Code)
		body := strings.TrimSpace(response.Body.String())
		var cafe []string
		if body != "" {
			cafe = strings.Split(body, ",")
		}
		assert.Len(t, cafe, v.want)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		parameters := url.Values{}
		parameters.Set("city", "moscow")
		parameters.Set("search", v.search)
		req := httptest.NewRequest("GET", "/cafe?"+parameters.Encode(), nil)
		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)
		body := strings.TrimSpace(response.Body.String())
		var cafe []string
		if body != "" {
			cafe = strings.Split(body, ",")
		}
		assert.Len(t, cafe, v.wantCount)
		for _, cf := range cafe {
			assert.Contains(t,
				strings.ToLower(cf),
				strings.ToLower(v.search),
			)
		}
	}
}
