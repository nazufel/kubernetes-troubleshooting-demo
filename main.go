package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", home)
	r.HandleFunc("/healthcheck", healthCheck)

	if err := http.ListenAndServe(":9999", r); err != nil {
		log.Fatalf("couldn't start server: %v\n", err)
	}

}

// environmentCheck checks for the existence of required env vars.
func environmentCheck() bool {

	// check for DEMO_YEAR env
	t := time.Now()

	val, ok := os.LookupEnv("DEMO_YEAR")
	if !ok {
		log.Println("did not find expected env var: 'DEMO_YEAR'")
		return false
	}

	if val != strconv.Itoa(t.Year()) {
		log.Printf("DEMO_YEAR env var did not match expected: %v, got: %v\n", t.Year(), val)
		return false
	}

	return true
}

// healthCheck returns the health check of the service to Kubernetes
func healthCheck(w http.ResponseWriter, _ *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	resp := make(map[string]string)

	// a healthy service has all required env vars.
	// checking they exist is part of a health check
	ok := environmentCheck()
	if !ok {

		resp["status"] = "unhealthy"
		resp["error"] = "missing required env vars"
		w.WriteHeader(http.StatusInternalServerError)
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("uanble to marshal json: %s\n", err)
		}
		w.Write(jsonResp)
		return
	}
	w.WriteHeader(http.StatusOK)

	resp["status"] = "healthy"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("uanble to marshal json: %s\n", err)
	}

	log.Println("hit healthckeck endpoint")

	w.Write(jsonResp)
}

// home returns with the application's primary message
func home(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	resp := make(map[string]string)
	resp["message"] = "hello K8s toubleshooting demo"
	resp["year"] = os.Getenv("DEMO_YEAR")
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("uanble to marshal json: %s", err)
	}
	w.Write(jsonResp)
}
