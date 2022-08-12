# Midata to Sheets

This program sources and sends UK bank [**midata**](https://www.hsbc.co.uk/current-accounts/midata-faqs/) from a directory straight to Google Sheets for further processing.

## Prerequisites

- Export midata in .CSV format to `data/midata`
- Paste your Google credentials file (`credentials.json`) into `data`
- Create a file called `spreadsheetConfig.json`:
  - Define the object as follows:
  ```javascript
  { "spreadsheet_id": "<your spreadsheet id goes here>" }
  ```

## Execution

```properties
go run main.go
```
