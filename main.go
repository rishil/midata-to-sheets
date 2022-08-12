package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type MiData struct {
	Date        string
	Type        string
	Merchant    string
	Transaction string
	Balance     string
}
type SpreadsheetConfig struct {
	SpreadsheetID string `json:"spreadsheet_id"`
}

type Files struct {
	Files []File
}

type File struct {
	Name []string
}

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func main() {
	if _, err := os.Stat("data/history.json"); err == nil {
		fmt.Printf("File exists\n")
	} else {
		fmt.Printf("File does not exist\n")

		historyFile, err := os.OpenFile("data/history.json", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}

		if _, err = historyFile.WriteString(string("")); err != nil {
			panic(err)
		}
		historyFile.Close()
	}

	miDataFiles, err := os.ReadDir("data/midata")
	if err != nil {
		log.Fatal("need midata dir")
	}

	historyFiles := File{}
	content, err := ioutil.ReadFile("data/history.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(content, &historyFiles)
	if err != nil {
		fmt.Println("Adding first entry")
	}

	for _, f := range miDataFiles {
		if !contains(historyFiles.Name, f.Name()) {
			historyFiles.Name = append(historyFiles.Name, f.Name())
			fmt.Println("Sending to Sheets: " + f.Name())
			readMiData(f.Name())
			time.Sleep(45 * time.Second)

			e := os.Remove("data/history.json")
			if e != nil {
			}
		}

	}
	historyFile, err := os.Create("data/history.json")
	if err != nil {
		panic(err)
	}
	out, err := json.Marshal(File{Name: historyFiles.Name})
	if err != nil {
		panic(err)
	}

	if _, err = historyFile.WriteString(string(out)); err == nil {
		if !strings.Contains(string(out), "null") {
			fmt.Println("history.json: " + string(out))
		} else {
			fmt.Println("No data")
		}
	}

}

func readMiData(csvPath string) {
	lines, err := ReadCsv("data/midata/" + csvPath)
	if err != nil {
		panic(err)
	}

	data := MiData{}
	// Loop through lines & turn into object
	for _, line := range lines {
		if len(line) > 3 {
			transacFloat, _ := strconv.ParseFloat(line[3], 8)

			data = MiData{
				Date:        line[0],
				Type:        line[1],
				Merchant:    line[2],
				Transaction: fmt.Sprintf("%v", math.Abs(transacFloat)),
				Balance:     line[4],
			}
			write(data)
		}
	}
}

func write(data MiData) {
	ctx := context.Background()
	b, err := ioutil.ReadFile("data/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("%v", err)
	}

	configContent, err := ioutil.ReadFile("data/spreadsheetConfig.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var configPayload SpreadsheetConfig
	err = json.Unmarshal(configContent, &configPayload)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	records := [][]interface{}{{data.Date, data.Type, data.Merchant, data.Transaction, data.Balance}}

	rb := &sheets.ValueRange{
		Values: records,
	}

	valueInputOption := "RAW"
	insertDataOption := "INSERT_ROWS"

	response, err := srv.Spreadsheets.Values.Append(configPayload.SpreadsheetID, "MiData", rb).ValueInputOption(valueInputOption).InsertDataOption(insertDataOption).Context(ctx).Do()
	if err != nil || response.HTTPStatusCode != 200 {
		panic(err)
	}
}

func ReadCsv(filename string) ([][]string, error) {
	// Open CSV file
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	// Read File into a Variable
	csvr := csv.NewReader(f)
	csvr.FieldsPerRecord = -1

	// skip first row
	if _, err := csvr.Read(); err != nil {
		panic(err)
	}

	lines, err := csvr.ReadAll()
	if err != nil {
		panic(err)
	}

	return lines, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "data/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
