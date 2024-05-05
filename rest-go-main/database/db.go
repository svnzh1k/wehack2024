package database

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func Init() (*sql.DB, error) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=123 dbname=wehack_db sslmode=disable")
	err = FullfillDataBase(db, "init.sql")
	if err != nil {
		return nil, err
	}
	return db, err
}

func FullfillDataBase(db *sql.DB, script string) error {
	file, err := os.Open(script)
	if err != nil {
		return errors.New("error opening a file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var q strings.Builder
	for scanner.Scan() {
		q.WriteString(scanner.Text())
	}

	statements := strings.Split(q.String(), ";")
	fmt.Println(statements)
	for i := 0; i < len(statements); i++ {
		ExecuteScript(statements[i], db)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
	return nil
}

func ExecuteScript(script string, db *sql.DB) {
	_, err := db.Exec(script)
	if err != nil {
		fmt.Println()
		fmt.Println("error", err)
		fmt.Println(script)
	}
}
