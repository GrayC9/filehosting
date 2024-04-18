package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	//ticker := time.NewTicker(time.Minute)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	//DeleteExpiredRecords()

	for {
		select {
		case <-ticker.C:
			DeleteExpiredRecords()
		}
	}
}

func DeleteExpiredRecords() {
	dbPassword, exists := os.LookupEnv("MARIADB_PASSWORD")
	if !exists {
		log.Fatal("No MARIADB_PASSWORD")
	}
	dbHost, exists := os.LookupEnv("MARIADB_HOST")
	if !exists {
		log.Fatal("No MARIADB_HOST")
	}
	var connectionString string = fmt.Sprintf("root:%s@tcp(%s:3306)/filehosting", dbPassword, dbHost)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	currentTime := time.Now()
	oneMinuteAgo := currentTime.Add(-time.Minute)
	rows, err := db.Query("select file_key from file_storage where start_time < ?", oneMinuteAgo)

	var fileKeys []string

	for rows.Next() {
		var fileKey string
		if err := rows.Scan(&fileKey); err != nil {
			log.Fatal(err)
			return
		}
		fileKeys = append(fileKeys, fileKey)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
		return
	}

	var placeholders []string
	var args []interface{}
	for _, key := range fileKeys {
		placeholders = append(placeholders, "?")
		args = append(args, key)
	}
	query := "DELETE FROM file_storage WHERE file_key IN (" + strings.Join(placeholders, ", ") + ")"

	placeholdersString := strings.Join(placeholders, ", ")

	query = "DELETE FROM file_storage WHERE file_key IN (" + placeholdersString + ")"

	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(args...)
	if err != nil {
		log.Fatal(err)
		return
	}

	directoryPath := "/home/teleg/key_from_israel/temp_test/"

	for _, fileKey := range fileKeys {
		filePath := directoryPath + fileKey
		err := os.RemoveAll(filePath)
		if err != nil {
			log.Printf("Ошибка удаления файла %s: %v", filePath, err)
		} else {
			log.Printf("Файл %s удален успешно.", filePath)
		}
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Удалено %d записей", numRows)
}
