package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Constants for the database connection and other settings
const (
	dbUser     = "mysql"       // Change to your database username
	dbPassword = "secret"      // Change to your database password
	dbName     = "marketing"   // Change to your database name
	batchSize  = 1000          // Change the number of records per batch
	totalCodes = 200000        // Change the total number of discount codes you want to generate
)

// Letters for generating the discount code
var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// Function to generate a random 5-character discount code
func generateCode(r *rand.Rand) string {
	b := make([]rune, 5)  // Code length is 5 characters
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]  // Randomly select characters from the letters
	}
	return string(b)
}

func main() {
	// Seed the random number generator with the current time
	randSource := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSource)

	// Database connection string. Customize if necessary (host, port, charset, etc.)
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbName)

	// Open a connection to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)  // Terminate if connection fails
	}
	defer db.Close()  // Ensure the connection is closed when done

	// Check if the connection is successful
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to database")

	// Store unique codes to avoid duplicates
	uniqueCodes := make(map[string]struct{})

	// Start a new transaction to insert data in batches
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Prepare SQL statement for inserting discount codes into the database
	stmt, err := tx.Prepare("INSERT INTO discount_codes (code, product_id, description, max_amount, discount_percent) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	startTime := time.Now()  // Record start time for performance tracking
	count := 0  // Initialize counter for the number of codes inserted

	// Loop to generate and insert discount codes
	for count < totalCodes {
		// Prepare a batch of values for insertion
		values := make([]interface{}, 0, batchSize*5)

		// Generate discount codes in batches
		for i := 0; i < batchSize && count < totalCodes; i++ {
			// Generate a unique discount code
			code := generateCode(r)

			// Skip if the code is already used
			if _, exists := uniqueCodes[code]; exists {
				continue
			}
			uniqueCodes[code] = struct{}{}  // Mark the code as used

			// Randomly generate product ID, description, max amount, and discount percent
			productID := r.Intn(1000) + 1  // Random product ID between 1 and 1000
			description := fmt.Sprintf("Discount for product %d", productID)  // Generate a description
			maxAmount := float64(r.Intn(100) + 10)  // Random max amount between 10 and 110
			discountPercent := r.Intn(50) + 1  // Random discount percentage between 1 and 50

			// Add the generated values to the batch
			values = append(values, code, productID, description, maxAmount, discountPercent)
			count++  // Increment the count of generated codes
		}

		// If there are values to insert, generate a query for batch insertion
		if len(values) > 0 {
			query := "INSERT INTO discount_codes (code, product_id, description, max_amount, discount_percent) VALUES " +
				strings.Repeat("(?, ?, ?, ?, ?),", len(values)/5)  // Create placeholders for values
			query = query[:len(query)-1]  // Remove the last comma

			// Execute the insert query with the generated values
			_, err := tx.Exec(query, values...)
			if err != nil {
				log.Fatal(err)  // Handle any error that occurs during insertion
			}
		}
	}

	// Commit the transaction to save changes
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	// Print the total number of inserted records and the time taken
	fmt.Printf("Inserted %d records in %v\n", totalCodes, time.Since(startTime))
}
