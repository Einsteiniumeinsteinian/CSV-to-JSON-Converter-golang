package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Verify arguments contain atleast 1 file. and at most 10 files.
func checkArguments(paths []string) ([]string, error) {
	if len(paths) < 1 || len(paths) > 10 {
		return nil, fmt.Errorf("Scripts expects at least one argument of path")
	}

	return paths, nil
}

func validatePath(paths []string) ([]string, error) {
	wg := sync.WaitGroup{}
	errChan := make(chan error)

	for _, path := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					errChan <- fmt.Errorf("%s File does not exist.\n", path)
				} else {
					errChan <- fmt.Errorf("Error: %v", err)
				}
			}
		}(path)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		return nil, err
	}

	return paths, nil
}

func validateExtensions(paths []string) ([]string, error) {
	for _, path := range paths {
		if filepath.Ext(path) != ".csv" {
			return nil, fmt.Errorf("%s has the wrong file extension\n", path)
		}
	}
	return paths, nil
}

// Function to remove the file extension
func removeFileExtension(path string) string {
	ext := filepath.Ext(path)
	return strings.TrimSuffix(path, ext)
}

func decodeFiles(file io.Reader, path string) {
	destinationPath := removeFileExtension(path) + ".json"
	scanner := bufio.NewScanner(file)
	output := make(chan string)
	var mutex sync.Mutex
	var jsonArray []map[string]string

	go func() {
		defer close(output)
		scanner.Scan()
		firstLine := scanner.Text()
		firstLineArray := convertStringToArray(firstLine)

		for scanner.Scan() {
			line := scanner.Text()
			lineArray := convertStringToArray(line)
			processLine(lineArray, firstLineArray, output, &mutex)
		}

	}()

	if err := scanner.Err(); err != nil {
		fmt.Println("Error: Something went wrong in scanner", err)
		os.Exit(411)
	}

	for line := range output {
		var jsonObject map[string]string
		err := json.Unmarshal([]byte(line), &jsonObject)

		if err != nil {
			fmt.Println("jSON Marshall Error:", err)
			os.Exit(412)
		}

		jsonArray = append(jsonArray, jsonObject)
	}

	if err := writeJSONToNewFile(jsonArray, destinationPath); err != nil {
		fmt.Println("Could not write JSON File:", err)
		os.Exit(413)
	}

	fmt.Println("JSON array written to ", destinationPath)
}

func processLine(values []string, keys []string, output chan<- string, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()
	output <- convertSliceToJSON(keys, values)
}

func writeJSONToNewFile(jsonArray []map[string]string, destinationPath string) error {
	// destinationPath := removeFileExtension(path) + ".json"
	newFile, err := os.Create(destinationPath)
	defer newFile.Close()

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Encode the JSON array and write to the new file
	encoder := json.NewEncoder(newFile)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(jsonArray)

	if err != nil {
		return err
	}

	return nil
}

func convertSliceToJSON(keys []string, values []string) string {
	data := make(map[string]interface{})

	// Iterate over the array and split each element by ":"
	for index, key := range keys {
		if key == "" {
			continue
		}
		data[key] = values[index]
	}

	// Convert the map to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("JSON Compile Error occurred:", err)
		os.Exit(20)
	}

	return string(jsonData)
}

func convertStringToArray(str string) []string {
	// Create a CSV reader
	reader := csv.NewReader(strings.NewReader(str))
	reader.TrimLeadingSpace = true
	record, err := reader.Read() // Read a single record
	
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return record
}

func compose(funcs ...func([]string) ([]string, error)) func([]string) ([]string, error) {
	return func(args []string) ([]string, error) {
		var err error

		for index, f := range funcs {
			args, err = f(args)
			if err != nil {
				fmt.Println(err)
				os.Exit(index + 1)
			}
		}

		return args, nil
	}
}

func main() {
	verifyArguments := compose(checkArguments, validatePath, validateExtensions)
	verifiedPaths, _ := verifyArguments(os.Args[1:])
	wg := sync.WaitGroup{}

	for _, path := range verifiedPaths {
		wg.Add(1)

		go func(path string) {
			file, err := os.Open(path)
			defer file.Close()

			if err != nil {
				fmt.Printf("issue with file: %s\n", err)
				os.Exit(41)
			}
			decodeFiles(file, path)

			wg.Done()

		}(path)
	}

	wg.Wait()
}
