package integration

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadBalancerIntegration(t *testing.T) {
	if _, exists := os.LookupEnv("ENABLE_INTEGRATION_TEST"); !exists {
		t.Skip("Integration test is not enabled")
	}

	client := http.Client{
		Timeout: 3 * time.Second,
	}
	baseAddress := "http://loadbalancer:8090"
	servers := []string{
		"server1:8080",
		"server2:8080",
		"server3:8080",
	}

	testIPs := []string{
		"192.168.1.1",
		"192.168.1.2",
		"192.168.1.3:8080",
		"192.168.1.4",
		"192.168.1.5:2121",
	}

	expectedBindings := map[string][]string{
		"server1:8080": {"192.168.1.2", "192.168.1.4", "192.168.1.5:2121"},
		"server2:8080": {"192.168.1.3:8080"},
		"server3:8080": {"192.168.1.1"},
	}

	getExpectedBinding := func(ip string) string {
		for _, server := range servers {
			if contains(expectedBindings[server], ip) {
				return server
			}
		}
		panic(fmt.Sprintf("cannot find binding for %s", ip))
	}

	var wg sync.WaitGroup
	wg.Add(len(testIPs))

	for _, ip := range testIPs {
		go func(ip string) {
			defer wg.Done()

			req, _ := http.NewRequest("GET", baseAddress, nil)
			req.Header.Set("X-Forwarded-For", ip)

			resp, fetchErr := client.Do(req)
			assert.Nil(t, fetchErr, "Error fetching response")
			defer resp.Body.Close()

			lbFrom := resp.Header.Get("lb-from")
			binding, found := expectedBindings[lbFrom]
			assert.True(t, found, "Unexpected lb-from header value: %s", lbFrom)

			isValid := contains(binding, ip)
			assert.True(t, isValid, "Expected %s to be in %v, got %v", ip, getExpectedBinding(ip), lbFrom)
		}(ip)
	}

	wg.Wait()
}

func BenchmarkLoadBalancer(b *testing.B) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	baseAddress := "http://loadbalancer:8090"

	parallel := 1000
	interval := time.Second
	total := 10

	testIPs := []string{
		"192.168.1.1",
		"192.168.1.2",
		"192.168.1.3",
	}

	var wg sync.WaitGroup
	wg.Add(parallel)

	start := make(chan struct{})

	benchmarks := make([][]time.Duration, parallel)

	for i := 0; i < parallel; i++ {
		req, _ := http.NewRequest("GET", baseAddress, nil)
		req.Header.Set("X-Forwarded-For", testIPs[i%len(testIPs)])

		go func(r *http.Request) {
			defer wg.Done()
			var (
				count     int
				durations = make([]time.Duration, total)
			)

			<-start

			for range time.Tick(interval) {
				start := time.Now()
				resp, fetchErr := client.Do(r)
				if fetchErr != nil {
					return
				}
				resp.Body.Close()
				durations[count] = time.Since(start)
				if count++; count == total {
					break
				}
			}

			benchmarks[i] = durations
		}(req)
	}

	close(start)

	wg.Wait()

	var result time.Duration

	for i := 0; i < total; i++ {
		var (
			sum   time.Duration
			count int
		)

		for j := 0; j < parallel; j++ {
			benchmark := benchmarks[j]
			if benchmark == nil {
				continue
			}
			res := benchmark[i]
			if res == 0 {
				continue
			}
			sum += res
			count++
		}

	}

	b.Logf("Average request duration: %v", time.Duration(result/time.Duration(total)))
}

func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}