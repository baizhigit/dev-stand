package healthcheck

import (
	"fmt"
	"net/http"
	"sync"
)

const success = "success"

type CheckFunc func() error

type result struct {
	name   string
	output string
}

func (h *handler) check(checks map[string]CheckFunc, out map[string]string) int {
	h.mu.RLock() // RLock: multiple concurrent probe requests run simultaneously
	defer h.mu.RUnlock()

	resultChan := make(chan result, len(checks)) // buffered = no goroutine leaks
	var wg sync.WaitGroup

	for name, check := range checks {
		wg.Add(1)
		go func(name string, check CheckFunc) {
			defer func() {
				// CRITICAL ORDER: recover and send result BEFORE wg.Done()
				// If wg.Done() runs first, the channel may be closed before
				// the panic path sends its result → send on closed channel panic
				if r := recover(); r != nil {
					resultChan <- result{name: name, output: fmt.Sprintf("panic: %v", r)}
				}
				wg.Done()
			}()
			output := success
			if err := check(); err != nil {
				output = err.Error()
			}
			resultChan <- result{name: name, output: output}
		}(name, check)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	status := http.StatusOK
	for res := range resultChan {
		out[res.name] = res.output
		if res.output != success {
			status = http.StatusServiceUnavailable
		}
	}
	return status
}
