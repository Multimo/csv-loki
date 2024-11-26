package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"
)

func main() {
	// csvFiles := []string{"./Prod3.csv", "./Prod1.csv", "./Prod2.csv"}
	// csvFiles := []string{"./Prod2.csv"}
	csvFiles := []string{"./Prod3.csv", "./Prod1.csv"}

	// Create a log file
	logFile := "./logs/processed_logs.log"
	logOutput, err := os.Create(logFile)
	if err != nil {
		log.Fatalf("Error creating log file: %v", err)
	}
	defer logOutput.Close()

	logger := slog.New(slog.NewTextHandler(logOutput, nil))

	for _, file := range csvFiles {
		fmt.Println("Processing CSV file:", file)
		logCSVLogs(logger, file)
		fmt.Println("Completed CSV file:", file)
	}

	log.Println("CSV processing completed")
}

func logCSVLogs(logger *slog.Logger, csvFile string) {
	file, err := os.Open(csvFile)
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		log.Fatalf("Error reading CSV headers: %v", err)
	}

	for {
		row, err := reader.Read()
		if err != nil {
			fmt.Println("Error reading CSV row:", err)
			break
		}

		record := make(map[string]string)
		for i, header := range headers {
			if i < len(row) {
				record[header] = row[i]
			}
		}

		timestamp := record["@timestamp"]

		ts := time.Now()
		if timestamp == "" {
			ts = time.Now()
		} else {
			// Nov 25, 2024 @ 11:38:36.066
			t, err := time.Parse("Jan 2, 2006 @ 15:04:05.000", timestamp)
			if err != nil {
				panic(err)
			}

			m := t.UnixMilli()
			t.Add(time.Duration(m * int64(time.Minute)))

			ts = t
		}

		level := record["level"]
		l := slog.LevelInfo
		if level == "" {
			l = slog.LevelInfo
		}
		if level == "error" {
			l = slog.LevelError
		}
		if level == "debug" {
			l = slog.LevelDebug
		}

		entry := slog.NewRecord(ts, l, record["msg"], 0)
		for key, value := range record {
			if value == "-" || value == "" {
				continue
			}

			entry.Add(strings.ReplaceAll(key, ".", "_"), value)
		}

		err = logger.Handler().Handle(context.TODO(), entry)
		if err != nil {
			// Handle error (optional)
			panic(err)
		}
	}

}
