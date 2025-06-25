package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// RequestResult holds the outcome of a single request
type RequestResult struct {
	Latency    time.Duration
	Successful bool
}

func main() {
	loadBalancerURL := "http://localhost:8000"
	outputDir := "images"

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	// Test scenarios: (RPS, Duration)
	scenarios := []struct {
		rps      int
		duration time.Duration
	}{
		{rps: 10, duration: 10 * time.Second},
		{rps: 50, duration: 10 * time.Second},
		{rps: 100, duration: 10 * time.Second},
		{rps: 200, duration: 10 * time.Second},
		{rps: 300, duration: 10 * time.Second},
		{rps: 500, duration: 10 * time.Second},
		{rps: 1000, duration: 10 * time.Second},
		{rps: 2000, duration: 10 * time.Second},
		{rps: 3000, duration: 10 * time.Second},
		{rps: 5000, duration: 10 * time.Second},
		{rps: 10000, duration: 10 * time.Second},
	}

	var allLatencies plotter.XYs
	var allRPSValues plotter.XYs
	var allAvailabilityValues plotter.XYs

	for _, scenario := range scenarios {
		fmt.Printf("\n--- Running Test Scenario: RPS = %d, Duration = %s ---\n", scenario.rps, scenario.duration)
		results := make(chan RequestResult, scenario.rps*int(scenario.duration.Seconds())*2) // Buffer channel
		var wg sync.WaitGroup

		ticker := time.NewTicker(time.Second / time.Duration(scenario.rps))
		done := make(chan struct{})

		go func() {
			for {
				select {
				case <-ticker.C:
					wg.Add(1)
					go func() {
						defer wg.Done()
						start := time.Now()
						resp, err := http.Get(loadBalancerURL)
						latency := time.Since(start)

						if err != nil {
							log.Printf("Request failed: %v", err)
							results <- RequestResult{Latency: latency, Successful: false}
							return
						}
						defer resp.Body.Close()

						if resp.StatusCode != http.StatusOK {
							bodyBytes, _ := ioutil.ReadAll(resp.Body)
							log.Printf("Request to %s failed with status %d: %s", loadBalancerURL, resp.StatusCode, string(bodyBytes))
							results <- RequestResult{Latency: latency, Successful: false}
							return
						}

						results <- RequestResult{Latency: latency, Successful: true}
					}()
				case <-done:
					ticker.Stop()
					return
				}
			}
		}()

		time.Sleep(scenario.duration)
		close(done)
		wg.Wait() // Wait for all ongoing requests to finish

		close(results) // Close the channel after all goroutines are done

		var successfulRequests int
		var totalRequests int
		var latencies []float64 // in milliseconds
		for res := range results {
			totalRequests++
			if res.Successful {
				successfulRequests++
				latencies = append(latencies, float64(res.Latency.Milliseconds()))
			}
		}

		if totalRequests == 0 {
			fmt.Println("No requests were made in this scenario.")
			continue
		}

		// Calculate Metrics
		rpsAchieved := float64(totalRequests) / scenario.duration.Seconds()
		availability := float64(successfulRequests) / float64(totalRequests) * 100

		sort.Float64s(latencies)

		p99Latency := 0.0
		if len(latencies) > 0 {
			p99Index := int(0.99 * float64(len(latencies)))
			if p99Index >= len(latencies) { // Handle edge case where p99 index might be out of bounds for small sets
				p99Index = len(latencies) - 1
			}
			p99Latency = latencies[p99Index]
		}

		fmt.Printf("  Total Requests: %d\n", totalRequests)
		fmt.Printf("  Successful Requests: %d\n", successfulRequests)
		fmt.Printf("  Achieved RPS: %.2f\n", rpsAchieved)
		fmt.Printf("  Availability: %.2f%%\n", availability)
		fmt.Printf("  P99 Latency: %.2f ms\n", p99Latency)

		allLatencies = append(allLatencies, plotter.XY{X: float64(scenario.rps), Y: p99Latency})
		allRPSValues = append(allRPSValues, plotter.XY{X: float64(scenario.rps), Y: rpsAchieved})
		allAvailabilityValues = append(allAvailabilityValues, plotter.XY{X: float64(scenario.rps), Y: availability})
	}

	// Plotting
	createPlot("P99 Latency vs. RPS", "Requests Per Second (Target)", "P99 Latency (ms)", allLatencies, outputDir+"/p99_latency.png")
	createPlot("Achieved RPS vs. Target RPS", "Requests Per Second (Target)", "Achieved Requests Per Second", allRPSValues, outputDir+"/achieved_rps.png")
	createPlot("Availability vs. RPS", "Requests Per Second (Target)", "Availability (%)", allAvailabilityValues, outputDir+"/availability.png")

	fmt.Println("\nTesting complete. Graphs saved in 'images/' folder.")
}

func createPlot(title, xLabel, yLabel string, data plotter.XYs, filename string) {
	p := plot.New()

	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	s, err := plotter.NewScatter(data)
	if err != nil {
		log.Fatal(err)
	}
	p.Add(s)

	l, err := plotter.NewLine(data)
	if err != nil {
		log.Fatal(err)
	}
	p.Add(l)

	p.Add(plotter.NewGrid())

	if err := p.Save(4*vg.Inch, 4*vg.Inch, filename); err != nil {
		log.Fatal(err)
	}
}
