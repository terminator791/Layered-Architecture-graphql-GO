package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/terminator791/Layered-Architecture-graphql-GO/proto"
)

type BenchmarkResult struct {
	TotalRequests    int
	SuccessfulReqs   int
	FailedReqs       int
	TotalDuration    time.Duration
	AvgResponseTime  time.Duration
	MinResponseTime  time.Duration
	MaxResponseTime  time.Duration
	RequestsPerSec   float64
}

func benchmarkHTTP(concurrency int, totalRequests int, userID string) *BenchmarkResult {
	results := make(chan time.Duration, totalRequests)
	errors := make(chan error, totalRequests)

	requestsPerWorker := totalRequests / concurrency
	remainder := totalRequests % concurrency

	start := time.Now()

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		workerRequests := requestsPerWorker
		if i < remainder {
			workerRequests++
		}

		go func(workerID int, reqCount int) {
			defer wg.Done()

			for j := 0; j < reqCount; j++ {
				reqStart := time.Now()

				// Create order request
				orderData := map[string]interface{}{
					"user_id":    userID,
					"product_id": fmt.Sprintf("product-%d-%d", workerID, j),
					"quantity":   1,
					"amount":     99.99,
				}

				jsonData, _ := json.Marshal(orderData)

				resp, err := http.Post("http://localhost:8080/orders",
					"application/json",
					bytes.NewBuffer(jsonData))

				reqDuration := time.Since(reqStart)

				if err != nil {
					errors <- err
				} else {
					// Read response body to ensure complete request
					io.ReadAll(resp.Body)
					resp.Body.Close()

					if resp.StatusCode == 200 {
						results <- reqDuration
					} else {
						errors <- fmt.Errorf("HTTP %d", resp.StatusCode)
					}
				}
			}
		}(i, workerRequests)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(results)
	close(errors)

	totalDuration := time.Since(start)

	// Collect results
	var durations []time.Duration
	successfulReqs := 0
	failedReqs := 0

	for duration := range results {
		durations = append(durations, duration)
		successfulReqs++
	}

	for range errors {
		failedReqs++
	}

	// Calculate statistics
	if len(durations) == 0 {
		return &BenchmarkResult{
			TotalRequests:  totalRequests,
			SuccessfulReqs: successfulReqs,
			FailedReqs:     failedReqs,
			TotalDuration:  totalDuration,
		}
	}

	minDuration := durations[0]
	maxDuration := durations[0]
	totalResponseTime := time.Duration(0)

	for _, duration := range durations {
		if duration < minDuration {
			minDuration = duration
		}
		if duration > maxDuration {
			maxDuration = duration
		}
		totalResponseTime += duration
	}

	avgResponseTime := totalResponseTime / time.Duration(len(durations))
	requestsPerSec := float64(successfulReqs) / totalDuration.Seconds()

	return &BenchmarkResult{
		TotalRequests:    totalRequests,
		SuccessfulReqs:   successfulReqs,
		FailedReqs:       failedReqs,
		TotalDuration:    totalDuration,
		AvgResponseTime:  avgResponseTime,
		MinResponseTime:  minDuration,
		MaxResponseTime:  maxDuration,
		RequestsPerSec:   requestsPerSec,
	}
}

func benchmarkGRPC(concurrency int, totalRequests int, userID string) *BenchmarkResult {
	// Connect to user service
	conn, err := grpc.Dial("localhost:9001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	// Prepare benchmark
	results := make(chan time.Duration, totalRequests)
	errors := make(chan error, totalRequests)

	requestsPerWorker := totalRequests / concurrency
	remainder := totalRequests % concurrency

	start := time.Now()

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		workerRequests := requestsPerWorker
		if i < remainder {
			workerRequests++
		}

		go func(workerID int, reqCount int) {
			defer wg.Done()

			for j := 0; j < reqCount; j++ {
				reqStart := time.Now()

				_, err := client.GetUser(context.Background(), &pb.GetUserRequest{
					Id: userID,
				})

				reqDuration := time.Since(reqStart)

				if err != nil {
					errors <- err
				} else {
					results <- reqDuration
				}
			}
		}(i, workerRequests)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(results)
	close(errors)

	totalDuration := time.Since(start)

	// Collect results
	var durations []time.Duration
	successfulReqs := 0
	failedReqs := 0

	for duration := range results {
		durations = append(durations, duration)
		successfulReqs++
	}

	for range errors {
		failedReqs++
	}

	// Calculate statistics
	if len(durations) == 0 {
		return &BenchmarkResult{
			TotalRequests:  totalRequests,
			SuccessfulReqs: successfulReqs,
			FailedReqs:     failedReqs,
			TotalDuration:  totalDuration,
		}
	}

	minDuration := durations[0]
	maxDuration := durations[0]
	totalResponseTime := time.Duration(0)

	for _, duration := range durations {
		if duration < minDuration {
			minDuration = duration
		}
		if duration > maxDuration {
			maxDuration = duration
		}
		totalResponseTime += duration
	}

	avgResponseTime := totalResponseTime / time.Duration(len(durations))
	requestsPerSec := float64(successfulReqs) / totalDuration.Seconds()

	return &BenchmarkResult{
		TotalRequests:    totalRequests,
		SuccessfulReqs:   successfulReqs,
		FailedReqs:       failedReqs,
		TotalDuration:    totalDuration,
		AvgResponseTime:  avgResponseTime,
		MinResponseTime:  minDuration,
		MaxResponseTime:  maxDuration,
		RequestsPerSec:   requestsPerSec,
	}
}

func printResults(protocol string, result *BenchmarkResult) {
	fmt.Printf("\n=== %s Performance Results ===\n", protocol)
	fmt.Printf("Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", result.SuccessfulReqs)
	fmt.Printf("Failed Requests: %d\n", result.FailedReqs)
	fmt.Printf("Total Duration: %v\n", result.TotalDuration)
	fmt.Printf("Average Response Time: %v\n", result.AvgResponseTime)
	fmt.Printf("Min Response Time: %v\n", result.MinResponseTime)
	fmt.Printf("Max Response Time: %v\n", result.MaxResponseTime)
	fmt.Printf("Requests per Second: %.2f\n", result.RequestsPerSec)
	fmt.Printf("Success Rate: %.2f%%\n", float64(result.SuccessfulReqs)/float64(result.TotalRequests)*100)
}

func main() {
	fmt.Println("🚀 gRPC vs HTTP Performance Comparison")
	fmt.Println("=====================================")

	// Create test user for gRPC tests
	conn, err := grpc.Dial("localhost:9001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)
	createResp, err := client.CreateUser(context.Background(), &pb.CreateUserRequest{
		Email: fmt.Sprintf("perf-test-%d@example.com", time.Now().Unix()),
		Name:  "Performance Test User",
	})
	if err != nil {
		log.Fatalf("Failed to create test user: %v", err)
	}

	userID := createResp.User.Id
	fmt.Printf("Created test user with ID: %s\n", userID)

	// Test scenarios
	scenarios := []struct {
		concurrency    int
		totalRequests  int
		description    string
	}{
		{50, 100, "Low Load: 5 concurrent, 100 requests"},
		{100, 500, "Medium Load: 10 concurrent, 500 requests"},
		{1000, 1000, "High Load: 20 concurrent, 1000 requests"},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n📊 Testing Scenario: %s\n", scenario.description)
		fmt.Printf("Concurrency: %d, Total Requests: %d\n", scenario.concurrency, scenario.totalRequests)

		// Test gRPC
		fmt.Println("\n🔹 Testing gRPC...")
		grpcResult := benchmarkGRPC(scenario.concurrency, scenario.totalRequests, userID)
		printResults("gRPC", grpcResult)

		// Test HTTP
		fmt.Println("\n🔹 Testing HTTP...")
		httpResult := benchmarkHTTP(scenario.concurrency, scenario.totalRequests, userID)
		printResults("HTTP", httpResult)

		// Comparison
		fmt.Println("\n⚡ Performance Comparison:")
		if grpcResult.RequestsPerSec > httpResult.RequestsPerSec {
			fmt.Printf("🏆 gRPC is %.1fx faster than HTTP\n",
				grpcResult.RequestsPerSec/httpResult.RequestsPerSec)
		} else {
			fmt.Printf("🏆 HTTP is %.1fx faster than gRPC\n",
				httpResult.RequestsPerSec/grpcResult.RequestsPerSec)
		}

		fmt.Printf("📈 gRPC Latency: %v vs HTTP Latency: %v\n",
			grpcResult.AvgResponseTime, httpResult.AvgResponseTime)

		// Brief pause between tests
		time.Sleep(3 * time.Second)
	}

	fmt.Println("\n✅ Performance comparison completed!")
}
