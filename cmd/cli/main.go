package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	user "github.com/terminator791/Layered-Architecture-graphql-GO/proto"
	"github.com/terminator791/Layered-Architecture-graphql-GO/pkg/models"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cli <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  create-user <email> <name>")
		fmt.Println("  get-user <id>")
		fmt.Println("  create-order <user_id> <product_id> <quantity> <amount>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create-user":
		if len(os.Args) != 4 {
			fmt.Println("Usage: cli create-user <email> <name>")
			os.Exit(1)
		}
		createUser(os.Args[2], os.Args[3])
	case "get-user":
		if len(os.Args) != 3 {
			fmt.Println("Usage: cli get-user <id>")
			os.Exit(1)
		}
		getUser(os.Args[2])
	case "create-order":
		if len(os.Args) != 6 {
			fmt.Println("Usage: cli create-order <user_id> <product_id> <quantity> <amount>")
			os.Exit(1)
		}
		createOrder(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func createUser(email, name string) {
	// Connect to user service via gRPC
	conn, err := grpc.NewClient("localhost:9001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to user service: %v", err)
	}
	defer conn.Close()

	client := user.NewUserServiceClient(conn)

	req := &user.CreateUserRequest{
		Email: email,
		Name:  name,
	}

	resp, err := client.CreateUser(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("User created successfully:\n")
	fmt.Printf("ID: %s\n", resp.User.Id)
	fmt.Printf("Email: %s\n", resp.User.Email)
	fmt.Printf("Name: %s\n", resp.User.Name)
}

func getUser(id string) {
	// Connect to user service via gRPC
	conn, err := grpc.NewClient("localhost:9001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to user service: %v", err)
	}
	defer conn.Close()

	client := user.NewUserServiceClient(conn)

	req := &user.GetUserRequest{
		Id: id,
	}

	resp, err := client.GetUser(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}

	fmt.Printf("User found:\n")
	fmt.Printf("ID: %s\n", resp.User.Id)
	fmt.Printf("Email: %s\n", resp.User.Email)
	fmt.Printf("Name: %s\n", resp.User.Name)
}

func createOrder(userID, productID, quantityStr, amountStr string) {
	// Parse quantity and amount
	var quantity int
	var amount float64
	if _, err := fmt.Sscanf(quantityStr, "%d", &quantity); err != nil {
		log.Fatalf("Invalid quantity: %v", err)
	}
	if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
		log.Fatalf("Invalid amount: %v", err)
	}

	// Create order via HTTP API to order service
	orderInput := models.CreateOrderInput{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Amount:    amount,
	}

	jsonData, err := json.Marshal(orderInput)
	if err != nil {
		log.Fatalf("Failed to marshal order input: %v", err)
	}

	resp, err := http.Post("http://localhost:8080/orders", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to create order: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error creating order: %s\n", string(body))
		os.Exit(1)
	}

	var order models.Order
	if err := json.Unmarshal(body, &order); err != nil {
		log.Fatalf("Failed to parse order response: %v", err)
	}

	fmt.Printf("Order created successfully:\n")
	fmt.Printf("ID: %s\n", order.ID)
	fmt.Printf("User ID: %s\n", order.UserID)
	fmt.Printf("Product ID: %s\n", order.ProductID)
	fmt.Printf("Quantity: %d\n", order.Quantity)
	fmt.Printf("Amount: %.2f\n", order.Amount)
	fmt.Printf("Status: %s\n", order.Status)
}