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

const (
	dbUser     = "mysql"
	dbPassword = "root"
	dbName     = "marketing"
	batchSize  = 1000
	totalCodes = 200000
)

var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func generateCode(r *rand.Rand) string {
	b := make([]rune, 5)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	randSource := rand.NewSource(time.Now().UnixNano())

	r := rand.New(randSource)
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to database")

	uniqueCodes := make(map[string]struct{})

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("INSERT INTO discount_codes (code, product_id, description, max_amount, discount_percent) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	startTime := time.Now()
	count := 0

	for count < totalCodes {
		values := make([]interface{}, 0, batchSize*5)

		for i := 0; i < batchSize && count < totalCodes; i++ {
			code := generateCode(r)

			if _, exists := uniqueCodes[code]; exists {
				continue
			}
			uniqueCodes[code] = struct{}{}

			productID := r.Intn(1000) + 1
			description := fmt.Sprintf("Discount for product %d", productID)
			maxAmount := float64(r.Intn(100) + 10)
			discountPercent := r.Intn(50) + 1

			values = append(values, code, productID, description, maxAmount, discountPercent)
			count++
		}

		if len(values) > 0 {
			query := "INSERT INTO discount_codes (code, product_id, description, max_amount, discount_percent) VALUES " +
				strings.Repeat("(?, ?, ?, ?, ?),", len(values)/5)
			query = query[:len(query)-1] 

			_, err := tx.Exec(query, values...)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Inserted %d records in %v\n", totalCodes, time.Since(startTime))
}