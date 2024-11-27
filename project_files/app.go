package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"path/filepath"

	"cloud.google.com/go/bigquery"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/theme"
	"google.golang.org/api/iterator"
)


func queryBasic(w io.Writer, projectID string, query_string string) error {
	

	/***************************
		 Set up BQ Client
	****************************/

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()


	q := client.Query(query_string)
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




/**************************
		App Theme		 
**************************/
type appTheme struct{}

func (t appTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 0x33, G: 0x61, B: 0xAC, A: 0xFF} // #3361AC
	case theme.ColorNameButton:
		return color.RGBA{R: 0xE8, G: 0xAF, B: 0x30, A: 0xFF} // #E8AF30
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 0x0F, G: 0x20, B: 0x43, A: 0xFF} // #0F2043
	case theme.ColorNameForeground:
		return color.White
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0xE8, G: 0xAF, B: 0x30, A: 0xFF} // #E8AF30
	case theme.ColorNameHover:
		return color.RGBA{R: 0xFF, G: 0xC1, B: 0x35, A: 0xFF} // Lighter button color for hover
	case theme.ColorNameDisabled:
		return color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xFF} // Gray for disabled elements
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t appTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t appTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t appTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 18
	default:
		return theme.DefaultTheme().Size(name)
	}
}




/********************************
			Start App
********************************/
func main() {
	
	myApp := app.New()
	myApp.Settings().SetTheme(&appTheme{}) 
	myWindow := myApp.NewWindow("GoQu BigQuery Export Tool")
	
	/******************************
			 Query Input
	******************************/
	text := canvas.NewText("Go Query", color.White)
	text.Alignment = fyne.TextAlignCenter
	text.TextStyle = fyne.TextStyle{Bold: true}

	
	ProjectID := widget.NewEntry()
	ProjectID.SetPlaceHolder("Enter Billing Project") 
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Write a Query")
	
	
	
	query_input := container.NewVBox(input, widget.NewButton("Run", func() {

		// Pass BQ API logic here
		err := queryBasic(os.Stdout, ProjectID.Text, input.Text)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		log.Println("Content was:", input.Text)

	
	}))

	query_input.Resize(fyne.NewSize(100, 40))
	

	content := container.NewVBox(
		text,
		ProjectID,
		query_input,
	)


	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.ShowAndRun()
}

