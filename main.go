// main.go
package main

import (
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

// represents a backend server
type Backend struct {
	URL          *url.URL
	Alive        atomic.Bool // for thread-safe boolean
	ReverseProxy *httputil.ReverseProxy
}

// ServerPool holds the collection of backends
type ServerPool struct {
	backends []*Backend
	current  uint64 // For Round Robin
}

// adds a new backend to the server pool
func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

// etNextBackendIndex returns the index of the next backend for Round Robin
func (s *ServerPool) GetNextBackendIndex() int {
	return int(atomic.AddUint64(&s.current, 1) % uint64(len(s.backends)))
}

// returns the index of a random backend
func (s *ServerPool) GetRandomBackendIndex() int {
	return rand.Intn(len(s.backends))
}

// GetBackend returns a backend based on the chosen algorithm
func (s *ServerPool) GetBackend(algorithm string) *Backend {
	var backend *Backend
	switch algorithm {
	case "round_robin":
		idx := s.GetNextBackendIndex()
		backend = s.backends[idx]
	case "random":
		idx := s.GetRandomBackendIndex()
		backend = s.backends[idx]
	default:
		log.Printf("Warning: Unknown algorithm '%s'. Defaulting to round_robin.\n", algorithm)
		idx := s.GetNextBackendIndex()
		backend = s.backends[idx]
	}
	return backend
}

// LoadBalancer is the HTTP handler for the load balancer
type LoadBalancer struct {
	pool      *ServerPool
	algorithm string
}

func NewLoadBalancer(algorithm string, backendURLs []string) *LoadBalancer {
	pool := &ServerPool{}
	for _, urlStr := range backendURLs {
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			log.Fatalf("Failed to parse backend URL %s: %v", urlStr, err)
		}
		backend := &Backend{
			URL:          parsedURL,
			ReverseProxy: httputil.NewSingleHostReverseProxy(parsedURL),
		}
		backend.Alive.Store(true) // Initially set all backends as alive
		pool.AddBackend(backend)
	}
	return &LoadBalancer{
		pool:      pool,
		algorithm: algorithm,
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.pool.GetBackend(lb.algorithm)
	if backend == nil || !backend.Alive.Load() {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	log.Printf("Proxying request to %s using %s algorithm\n", backend.URL.String(), lb.algorithm)
	backend.ReverseProxy.ServeHTTP(w, r)
}

func main() {

	// Backend servers the load balancer will distribute requests to
	backendURLs := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	// Choose your load balancing algorithm
	// "round_robin", "random"
	chosenAlgorithm := "round_robin"
	// chosenAlgorithm := "random"

	lb := NewLoadBalancer(chosenAlgorithm, backendURLs)

	log.Printf("Starting Load Balancer on :8000 with %s algorithm\n", chosenAlgorithm)
	http.Handle("/", lb)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("Load balancer failed: %v", err)
	}
}
