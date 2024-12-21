package bigQuery

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	oathCredentialFile = "/home/eyelady/projects/go/bq/secrets/oathtoken.json"
)

func openBrowserForAuth(url string) error {
	var cmd string
	var args []string
	
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


/**********************************************************************
	Manages Browser Pop up for OAuth 2 authentication.
	I utilize the local URI callback approach for desktop applications.
***********************************************************************/
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	
	// Create a Listener for random port and assign it at runtime.
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("Unable to create listener: %v", err)
	}
	defer listener.Close()

	
	port := listener.Addr().(*net.TCPAddr).Port
	config.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", port)
	

	state := fmt.Sprintf("%d", time.Now().UnixNano())
	codeChan := make(chan string)
	var server *http.Server


	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}
		
		code := r.URL.Query().Get("code")
		codeChan <- code
		
		// Actual design for redirect url page.
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w,
		`<!DOCTYPE html>
		<html lang="en">
	
		<body style="background-color: #3361AC";>
			<h1 style="color:#E8AF30; font-family='Verdana'; text-align: center;"> Authentication Successful </h1>
			<h2 style="color:white; font-family='Verdana'; text-align: center;"> 
			Your query will write to a file called query_result.csv in your downloads folder.
			You can close this window now. 
			</h2>
		</body>

		</html>
		`)
		

		go func() {
			server.Shutdown(context.Background())
		}()
	})


	server = &http.Server{Handler: http.DefaultServeMux}
	go server.Serve(listener)

	
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	err = openBrowserForAuth(authURL)
	if err != nil {
		log.Printf("Unable to open browser: %v", err)
		log.Printf("Please open this URL manually: %s", authURL)
	}

	code := <-codeChan
	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to exchange code for token: %v", err)
	}

	return tok
}




// Token Variable
var (
	oauthConfig *oauth2.Config
	once sync.Once
)


/****************************************************************
    his sets the scope for the app on requested permissions
****************************************************************/
func getOAuthConfig() *oauth2.Config {
	once.Do(func() {
		b, err := os.ReadFile(oathCredentialFile)
		if err != nil {
			log.Fatalf("Error finding OAuth token: %v", err)
		}
		
		config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/bigquery")
		if err != nil {
			log.Fatalf("Error Parsing OAuth Token: %v", err)
		}
		oauthConfig = config
	})
	return oauthConfig
}



/********************************************************************
            Handles our actual job context and params
********************************************************************/
func RunQueryJob(projectID string, query_string string) (*bigquery.Job, error) {
	ctx := context.Background()
	config := getOAuthConfig()
	appToken := getTokenFromWeb(config)
	
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
