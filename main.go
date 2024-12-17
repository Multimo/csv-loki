package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// csvFiles := []string{"./Prod3.csv", "./Prod1.csv", "./Prod2.csv"}
	csvFiles := []string{"./Prod2-dec2.csv"}
	// csvFiles := []string{"./Prod7.csv"}

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

	from := 0
	// from := 100000
	limit := 100000

	counter := 0
	for {

		if counter == limit {
			break
		}

		counter++

		row, err := reader.Read()
		if err != nil {
			fmt.Println("Error reading CSV row:", err, row)
			break
		}

		if from > 0 && counter < from {
			continue
		}

		record := make(map[string]string)
		for i, header := range headers {
			if i < len(row) {
				record[header] = row[i]
			}
		}

		// log := make(map[string]interface{})
		if record["log"] != "" {
			// err := json.Unmarshal([]byte(record["log"]), &log)
			// if err != nil {
			// 	fmt.Println("Error unmarshalling log field:", err)
			// 	fmt.Println("Error unmarshalling log field:", record["log"])
			// 	panic("Error unmarshalling log field")
			// }

			logmap, err := jsonToMap(record["log"])
			if err != nil {
				fmt.Println("Error converting log field to map:", err)
				fmt.Println("Error converting log field to map:", record["log"])
				panic("Error converting log field to map")
			}
			for key, value := range logmap {
				record[key] = value
			}
			record["log"] = ""
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

func flattenJSON(data map[string]interface{}, prefix string, result map[string]string) {
	for key, value := range data {
		// Construct the full key with dot notation
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// Handle the type of the value
		switch v := value.(type) {
		case string:
			result[fullKey] = v
		case float64:
			result[fullKey] = strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			result[fullKey] = strconv.FormatBool(v)
		case map[string]interface{}:
			// Recursively flatten nested objects
			flattenJSON(v, fullKey, result)
		case []interface{}:
			// Convert arrays to string representations
			for i, elem := range v {
				flattenJSON(map[string]interface{}{fmt.Sprintf("%s[%d]", fullKey, i): elem}, "", result)
			}
		default:
			// Convert other types to strings
			result[fullKey] = fmt.Sprintf("%v", v)
		}
	}
}

func jsonToMap(inputJSON string) (map[string]string, error) {
	// Parse JSON into a generic map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(inputJSON), &data); err != nil {
		return nil, err
	}

	// Create a result map to hold the flattened JSON
	result := make(map[string]string)
	flattenJSON(data, "", result)
	return result, nil
}
