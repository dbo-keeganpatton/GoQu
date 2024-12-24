package csvWriter

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)



// Progress Bar callback
type ProgressCallBack func(float64)


// Thread Safe CSV writer stuff
type SafeWriter struct {
	mutex sync.Mutex
	writer *csv.Writer
}

func (s *SafeWriter) Write(record []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.writer.Write(record)
}

func (s *SafeWriter) Flush() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.writer.Flush()
}



/* Accepts a *bigquery.Job struct as an argument and writes to a csv file. */
func WriteCsv(job *bigquery.Job, progressCb ProgressCallBack) error {
	
	
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
	if err != nil {
		return fmt.Errorf("Error creating file: %v", err)
	}
	defer file.Close()
	

	writer := csv.NewWriter(file)
	writeSafe := &SafeWriter{writer: writer}
	defer writeSafe.Flush()
		

	header := make([]string, 0)
	for _, field := range it.Schema {
		header = append(header, field.Name)
	}
	
	
	err = writeSafe.Write(header)
	if err != nil {
		log.Fatal("Error writing header:", err)
	}
	


	// GoRoutine stuff
	rowCount := it.TotalRows	
	rowsChan := make(chan []bigquery.Value)
	errChan := make(chan error)
	var wg sync.WaitGroup

	workers := 8
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for row := range rowsChan {
				csvRow := make([]string, len(row))
				for i, value := range row {
					csvRow[i] = fmt.Sprintf("%v", value)
				}
				if err := writeSafe.Write(csvRow); err != nil {
					errChan <- fmt.Errorf("error writing to csv: %v", err)
					return
				}
			}
		}()
	}



	// Main Iterator
	var rowsWritten int64 = 0
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("iterator.Next: %v", err)
		}

		// Send curr row to the curr channel
		rowsChan <- row
		rowsWritten++
		if rowCount > 0 {
			progress := float64(rowsWritten) / float64(rowCount)
			progressCb(progress)
		}
	}

	close(rowsChan)
	wg.Wait()

	select {
	case err := <- errChan:
		return err
	default:
	}
		
	log.Println("Success")
	return nil


}
