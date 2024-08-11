package readers

import (
	"fmt"
)

func SendEvents(howMany int) <-chan []byte {
	rowChan := make(chan []byte)

	go func() {
		defer close(rowChan)
		for i := 0; i <= howMany; i++ {
			rowChan <- []byte(fmt.Sprintf(`{"event": %d}`, i))
		}
	}()

	return rowChan
}
