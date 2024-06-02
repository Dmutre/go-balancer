package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Dmutre/go-balancer/httptools"
	"github.com/Dmutre/go-balancer/signal"
)

var port = flag.Int("port", 8080, "server port")

const confResponseDelaySec = "CONF_RESPONSE_DELAY_SEC"
const confHealthFailure = "CONF_HEALTH_FAILURE"

func main() {
	flag.Parse()

	h := new(http.ServeMux)

	h.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("content-type", "text/plain")
		if failConfig := os.Getenv(confHealthFailure); failConfig == "true" {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte("FAILURE"))
		} else {
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte("OK"))
		}
	})

	teamName := "KanchyEnjoyers"
	currentDate := time.Now().Format("2006-01-02")
	dbURL := "http://localhost:8081/db/" + teamName

	_, err := http.Post(dbURL, "application/json", strings.NewReader(fmt.Sprintf(`{"value": "%s"}`, currentDate)))
	if err != nil {
		panic(err)
	}

	h.HandleFunc("/api/v1/some-data", func(rw http.ResponseWriter, r *http.Request) {
		respDelayString := os.Getenv(confResponseDelaySec)
		if delaySec, parseErr := strconv.Atoi(respDelayString); parseErr == nil && delaySec > 0 && delaySec < 300 {
			time.Sleep(time.Duration(delaySec) * time.Second)
		}

		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(rw, "missing key parameter", http.StatusBadRequest)
			return
		}

		resp, err := http.Get("http://localhost:8081/db/" + key)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			http.NotFound(rw, r)
			return
		}

		if resp.StatusCode != http.StatusOK {
			http.Error(rw, "error from db service", http.StatusInternalServerError)
			return
		}

		var data map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("content-type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(data)
	})

	server := httptools.CreateServer(*port, h)
	server.Start()
	signal.WaitForTerminationSignal()
}
