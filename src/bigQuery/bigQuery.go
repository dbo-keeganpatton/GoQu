package bigQuery

import (
	"fmt"
	"context"
	"cloud.google.com/go/bigquery"
)



// Establishes a BigQuery client and executes a query string as input to return a BQ job.
func RunQueryJob(projectID string, query_string string) (*bigquery.Job, error) {
	

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	
	q := client.Query(query_string)
	q.Location = "US"

	
	job, err := q.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("query.Run: %v", err)
	}
	

	status, err := job.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("job.Wait: %v", err)
	}

	
	if err := status.Err(); err != nil {
		return nil, fmt.Errorf("job failed with error: %v", err)
	}
	

	return job, nil

}
