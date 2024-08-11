package readers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ReadCSVWithHeader(filePath string) (<-chan []byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(file)
	rowChan := make(chan []byte)

	go func() {
		defer close(rowChan)
		defer file.Close()

		header, err := reader.Read()
		if err != nil {
			fmt.Println("Error reading CSV header:", err)
			return
		}

		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("Error reading CSV row:", err)
				break
			}

			rowMap := make(map[string]string)
			for i, field := range row {
				rowMap[header[i]] = field
			}

			rowBytes, err := json.Marshal(rowMap)
			if err != nil {
				fmt.Println("Error marshalling JSON:", err)
				break
			}

			rowChan <- rowBytes
		}
	}()

	return rowChan, nil
}
