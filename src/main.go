package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Get service name from command line argument or environment
	serviceName := getServiceName()

	// Map of available services
	services := map[string]string{
		"api-gateway":             "./src/cmd/api-gateway/main.go",
		"content-service":         "./src/cmd/content-service/main.go",
		"decision-service":        "./src/cmd/decision-service/main.go",
		"hr-service":              "./src/cmd/hr-service/main.go",
		"financial-service":       "./src/cmd/financial-service/main.go",
		"governance-service":      "./src/cmd/governance-service/main.go",
		"legal-service":           "./src/cmd/legal-service/main.go",
		"risk-service":            "./src/cmd/risk-service/main.go",
		"self-improvement-service": "./src/cmd/self-improvement-service/main.go",
	}

	if serviceName == "help" || serviceName == "--help" || serviceName == "-h" {
		printUsage(services)
		return
	}

	// Default to API Gateway for backward compatibility
	if serviceName == "" {
		serviceName = "api-gateway"
	}

	serviceMain, exists := services[serviceName]
	if !exists {
		log.Fatalf("Unknown service: %s. Use 'help' to see available services.", serviceName)
	}

	log.Printf("Starting %s...", serviceName)

	// Execute the service
	cmd := exec.Command("go", "run", serviceMain)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to start service %s: %v", serviceName, err)
	}
}

// getServiceName gets the service name from command line args or environment
func getServiceName() string {
	// Check command line arguments first
	if len(os.Args) > 1 {
		return strings.ToLower(os.Args[1])
	}

	// Check environment variable
	if service := os.Getenv("SERVICE_NAME"); service != "" {
		return strings.ToLower(service)
	}

	return ""
}

// printUsage prints usage information
func printUsage(services map[string]string) {
	fmt.Println("Autonomous Content Service - Microservices Launcher")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run main.go [service-name]")
	fmt.Println("  SERVICE_NAME=service-name go run main.go")
	fmt.Println("")
	fmt.Println("Available services:")
	for service := range services {
		fmt.Printf("  - %s\n", service)
	}
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run main.go api-gateway")
	fmt.Println("  go run main.go content-service")
	fmt.Println("  SERVICE_NAME=hr-service go run main.go")
	fmt.Println("")
	fmt.Println("Default: api-gateway (for backward compatibility)")
}
