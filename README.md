# CSV TO JSON FILE CONVERTER

Can be used to convert csv files to a JSON file. It uses the first entires as the object key.

## Usage

The Scripts does the following

- checks arguments to ensure that csv file is passesd properly
- converts the csv file to a JSON file.
- converts up to 10 files at once

## Dependecies

- go programming language installed

clone repo

```go
   go run convert-csv-json.go <filename/filepath>
```

**Atleast one file must be provided as argument and must have extension of csv**. To convert multiple csv files run same command with multiple file path.

samples files have been provided for test:

``` go
go run convert-csv-json.go adjustments.csv new-top-surnames.csv state-pop.csv
```
