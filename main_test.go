package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
)

// TestHealthCheck tests the healthCheck endpoint for 200 status code
func TestHealthCheck(t *testing.T) {

	t.Run("returns health status", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/healthCheck", nil)
		response := httptest.NewRecorder()

		os.Setenv("DEMO_YEAR", "2022")
		healthCheck(response, request)

		got := response.Result().StatusCode
		want := http.StatusOK

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("test returned body", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/healthCheck", nil)
		response := httptest.NewRecorder()

		os.Setenv("DEMO_YEAR", "2022")
		healthCheck(response, request)

		wantRespBody := make(map[string]string)
		wantRespBody["status"] = "healthy"

		got := response.Body.Bytes()

		gotRespBody := make(map[string]string)

		err := json.Unmarshal([]byte(got), &gotRespBody)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}

		compare := reflect.DeepEqual(gotRespBody, wantRespBody)
		if !compare {
			t.Errorf("got %v, want %v", gotRespBody, wantRespBody)
		}
	})
}

// TestHome tests the home handler for a 200 status code
func TestHome(t *testing.T) {

	t.Run("test status code", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		home(response, request)

		got := response.Result().StatusCode

		want := http.StatusOK

		// make sure the returned status code is 200
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("test returned body", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		home(response, request)

		wantRespBody := make(map[string]string)
		wantRespBody["message"] = "hello K8s toubleshooting demo"
		wantRespBody["year"] = os.Getenv("DEMO_YEAR")

		got := response.Body.Bytes()

		gotRespBody := make(map[string]string)

		err := json.Unmarshal([]byte(got), &gotRespBody)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}

		compare := reflect.DeepEqual(gotRespBody, wantRespBody)
		if !compare {
			t.Errorf("got %v, want %v", gotRespBody, wantRespBody)
		}
	})
}

// TestRouter tests the routes of the http router
func TestRouter(t *testing.T) {

	t.Run("test home route status", func(t *testing.T) {

		r := mux.NewRouter()

		r.HandleFunc("/", home)

		req, _ := http.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("wrong status. got: %v, wanted:%v", rr.Code, http.StatusOK)
		}
	})

	t.Run("test healthCheck route status", func(t *testing.T) {

		r := mux.NewRouter()

		r.HandleFunc("/healthCheck", healthCheck)

		req, _ := http.NewRequest("GET", "/healthCheck", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("wrong status. got: %v, wanted:%v", rr.Code, http.StatusOK)
		}
	})

	t.Run("test bad route status", func(t *testing.T) {

		r := mux.NewRouter()

		req, _ := http.NewRequest("GET", "/foo", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("wrong status. got: %v, wanted:%v", rr.Code, http.StatusNotFound)
		}
	})
}

// TestEnvironmentCheck tests the func that checks for an env var populated by a ConfigMap
func TestEnvironmentCheck(t *testing.T) {

	t.Run("test env exists", func(t *testing.T) {

		os.Setenv("DEMO_YEAR", "2022")
		got := environmentCheck()
		want := true

		if got != want {
			t.Errorf("required environment variables not set")
		}
	})

	t.Run("test wrong env", func(t *testing.T) {

		os.Setenv("DEMO_YEAR", "2021")
		got := environmentCheck()
		want := false

		if got != want {
			t.Errorf("required environment variables not set")
		}
	})

	t.Run("test no env", func(t *testing.T) {

		got := environmentCheck()
		want := false

		if got != want {
			t.Errorf("required environment variables not set")
		}
	})
}
