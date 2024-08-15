package readers

import (
	"context"
	"fmt"
)

// SendEvents sends events down a []byte channel - can be cancelled via a context
func SendEvents(ctx context.Context, howMany int) <-chan []byte {
	rowChan := make(chan []byte)

	go func() {
		defer close(rowChan)

		for i := 0; i <= howMany; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				// continue with processing
			}

			rowChan <- []byte(fmt.Sprintf(`{"event": %d}`, i))
		}
	}()

	return rowChan
}
