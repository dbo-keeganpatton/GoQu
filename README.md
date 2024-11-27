# GoQu 
## Golang BigQuery Export Tool

### Mac Users
- **Download the GoQu_Mac file located the the Mac_File directory.

### Windows Users
- **Download the GoQu_Windows.exe file located in the Windows_File directory.

![GoQu Image](./images/GoQu.png)


# Guide for Use
Export your complete Query result to a CSV file, quickly and efficiently. 

Authentication is handled by setting the *GOOGLE_APPLICATION_CREDENTIALS* environment variable using the [gcloud cli](https://cloud.google.com/sdk/docs/install).


### Authentication
Authenticate by inputting the following command in your terminal, this will prompt a login to occur in your browser, ==select yes== for the cli to set your credentials variable automatically.

`
gcloud auth application-default login
`


### Use
Input your query as a string, as well as the billing project ID that will be used to bill for compute resources consumed by the query. 



## ToDo Feature List
- Query error parsing
- Syntax highlighting
