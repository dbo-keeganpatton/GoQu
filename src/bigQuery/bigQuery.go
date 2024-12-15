/* TO DO
!
! Implement Local http server for callback after auth option
!
*/

package bigQuery

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)


const (oathCredentialFile = "/home/eyelady/projects/go/bq/secrets/oathtoken.json")



/* This function invokes authentication via a pop up Browser 
   Without requiring direct clicking by the user. */
func openBrowserForAuth(url string) error {
	var cmd string
	var args []string
	
	// OS Commands for Invokation
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}

	case "darwin":
		cmd = "open"

	default: // Linux
		cmd = "xdg-open"	
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()

}



/* This Function handles Oauth2 for User specific scope. */
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	
	/*********************
	!	HTTP Callback	 !
	*********************/
	state := fmt.Sprintf("%d", time.Now().UnixNano())
	codeChan := make(chan string)
	
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
        }
	
	code := r.URL.Query().Get("code")
	codeChan <- code

	fmt.Fprintf(w, "Signed In, close window now.")

	})

	server := &http.Server{Addr: ":0"}
    go server.ListenAndServe()
	
	port := server.Addr
    if port == ":0" {
		port = ":80"     
	}
    
	config.RedirectURL = fmt.Sprintf("http://localhost%s/callback", port)



	/***********************
	!	Main Auth Logic    !
	***********************/
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	err := openBrowserForAuth(authURL)
	if err != nil {
		log.Printf("Cant get this dang browser open!: %v", err)
	}

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




/* Establishes a BigQuery client and executes a query string as input to return a BQ job. */
func RunQueryJob(projectID string, query_string string) (*bigquery.Job, error) {
	

	ctx := context.Background()
	
	// App Auth
	b, err := os.ReadFile(oathCredentialFile)
	if err != nil {
		log.Fatalf("Cannot find Oath Client Token: %v", err)
	}

	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/bigquery.readonly")
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
