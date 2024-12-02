package csvWriter

import (
	"os"
	"log"
	"fmt"
	"context"
	"encoding/csv"
	"path/filepath"
	"google.golang.org/api/iterator"
	"cloud.google.com/go/bigquery"
)



/* Accepts a *bigquery.Job struct as an argument and writes to a csv file. */
func WriteCsv(job *bigquery.Job) error {
	
	
	// Get the Job first
	ctx := context.Background()
	it, err := job.Read(ctx)
	if err != nil {
		return fmt.Errorf("failure to read job: %v", err)
	}


	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory:", err)
	}


	downloadPath := filepath.Join(homeDir, "Downloads", "query_result.csv")
	file, err := os.Create(downloadPath)

	
	writer := csv.NewWriter(file)
	defer writer.Flush()
		

	header := make([]string, 0)
	for _, field := range it.Schema {
		header = append(header, field.Name)
	}
	
	
	err = writer.Write(header)
	if err != nil {
		log.Fatal("Error writing header:", err)
	}


	// Main Iterator
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("iterator.Next: %v", err)
		}


		csvRow := make([]string, len(row))
		for i, value := range row {
			csvRow[i] = fmt.Sprintf("%v", value)
		}
		if err := writer.Write(csvRow); err != nil {
			return fmt.Errorf("error writing to csv: %v", err)
		}

	}	
	
	
	log.Println("Success")
	return nil


}
