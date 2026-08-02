package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("Financial Advisor core service")
	fmt.Println("Database URL:", os.Getenv("DATABASE_URL"))
	log.Println("Core service initialized")
}
