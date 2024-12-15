package bigQuery

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)


const ( oathCredentialFile = "/home/eyelady/projects/go/bq/secrets/oathtoken.json" )



func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Follow Link to Authenticate, and type code"+": \n%v\n", authURL)
	
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil{
		log.Fatalf("Unable to get token: %v", err)
	}

	return tok

}




// Establishes a BigQuery client and executes a query string as input to return a BQ job.
func RunQueryJob(projectID string, query_string string) (*bigquery.Job, error) {
	

	ctx := context.Background()
	
	// App Auth
	b, err := os.ReadFile(oathCredentialFile)
	if err != nil {
		log.Fatalf("Cannot find Oath Client Token: %v", err)
	}

	config, err := google.ConfigFromJSON(b, bigquery.Scope)
	if err != nil {
		log.Fatalf("Issue parsing Oath Client File: %v", err)
	}
	
	
	appToken := getTokenFromWeb(config)


	// Client stuff
	client, err := bigquery.NewClient(
		ctx, 
		projectID, 
		option.WithTokenSource(
			config.TokenSource(ctx, appToken)))
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
