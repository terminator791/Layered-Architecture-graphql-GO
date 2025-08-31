package main

import (
	"context"
	"fmt"
	"log"
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

func benchmarkGRPC(concurrency int, totalRequests int) *BenchmarkResult {
	// Connect to user service
	conn, err := grpc.Dial("localhost:9001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	// Create test user first
	createReq := &pb.CreateUserRequest{
		Email: fmt.Sprintf("benchmark-%d@example.com", time.Now().Unix()),
		Name:  "Benchmark User",
	}

	createResp, err := client.CreateUser(context.Background(), createReq)
	if err != nil {
		log.Fatalf("Failed to create test user: %v", err)
	}

	userID := createResp.User.Id
	fmt.Printf("Created test user with ID: %s\n", userID)

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

func printResults(result *BenchmarkResult) {
	fmt.Println("\n=== gRPC Performance Benchmark Results ===")
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
	fmt.Println("🚀 Starting gRPC Performance Benchmark")
	fmt.Println("=====================================")

	// Test scenarios
	scenarios := []struct {
		concurrency    int
		totalRequests  int
		description    string
	}{
		{10, 100, "Single thread, 100 requests"},
		{50, 500, "5 concurrent threads, 500 requests"},
		{100, 1000, "10 concurrent threads, 1000 requests"},
		{1000, 2000, "20 concurrent threads, 2000 requests"},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n📊 Testing: %s\n", scenario.description)
		fmt.Printf("Concurrency: %d, Total Requests: %d\n", scenario.concurrency, scenario.totalRequests)

		result := benchmarkGRPC(scenario.concurrency, scenario.totalRequests)
		printResults(result)

		// Brief pause between tests
		time.Sleep(2 * time.Second)
	}

	fmt.Println("\n✅ Benchmark completed!")
}
