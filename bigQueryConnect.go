package main

import (
	"encoding/csv"
	"context"
	"fmt"
	"log"
	"io"
	"os"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	
)




func queryBasic(w io.Writer, projectID string) error {
	

	/***************************
		 Set up BQ Client
	****************************/

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()


	q := client.Query(
		`select 
		occupation_title
		from healthcare-111-391317.hc_db_prod_111.hc_decade_projections`)
	q.Location = "US"
	job, err := q.Run(ctx)
	if err != nil {
		return err
	}
	

	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	if err := status.Err(); err != nil {
		return err
	}
	it, err := job.Read(ctx)
	

	/***************************
			Stage CSV
	****************************/
	file, err := os.Create("data.csv")
		if err != nil {
			log.Fatal("Error creating file:", err)
		}
		defer file.Close()

	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	header := []string{"job"}
	err = writer.Write(header)
	if err != nil {
		log.Fatal("Error writing header:", err)
	}

	
	/***************************
			Write CSV
	****************************/
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("iterator.Next: %v", err)
		}


		csvRow := []string{fmt.Sprintf("%v", row[0])}
		if err := writer.Write(csvRow); err != nil {
			return fmt.Errorf("error writing to csv: %v", err)
		}
	}	
	
	log.Println("Success")
	return nil
}



func main() {
	
	projectID := "healthcare-111-391317" 
	err := queryBasic(os.Stdout, projectID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}


}
