# Personal Finance

This program sources and sends UK bank [**midata**](https://www.hsbc.co.uk/current-accounts/midata-faqs/) from a directory straight to Google Sheets for further processing.

## Prerequisites

- Export MiData in .CSV format to `/midata`
- Paste your Google credentials file (`credentials.json`) in `data`
- Create a file called `spreadsheetConfig.json`:
  - Define the object as follows:
  ```javascript
  { "spreadsheet_id": "<your spreadsheet id goes here>" }
  ```

## Execution

- `go run main.go`
