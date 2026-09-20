//go:build e2e

package e2e

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
)

var (
	baseURL string

	lifecycleMu sync.Mutex
	finished    = map[string]bool{}
)

func TestMain(m *testing.M) {
	baseURL = os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	log.Printf("e2e: base URL is %s", baseURL)

	if err := waitForAPI(baseURL); err != nil {
		log.Printf("e2e API is not reachable at %s: %v\nstart the stack with `just e2e-up` first", baseURL, err)
		os.Exit(1)
	}

	started := time.Now()
	exitCode := m.Run()
	log.Printf("e2e: suite finished in %s (exit %d), run with `-v` for per-request detail", time.Since(started), exitCode)

	os.Exit(exitCode)
}

func waitForAPI(baseURL string) error {
	deadline := time.Now().Add(60 * time.Second)

	for time.Now().Before(deadline) {
		response, err := http.Get(baseURL + "/metrics")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timed out waiting for API")
}

func newExpect(t *testing.T) *httpexpect.Expect {
	t.Helper()

	logTestLifecycle(t)

	return httpexpect.WithConfig(httpexpect.Config{
		TestName: t.Name(),
		BaseURL:  baseURL,
		Client:   &http.Client{Timeout: 15 * time.Second},
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}

func logTestLifecycle(t *testing.T) {
	t.Helper()

	lifecycleMu.Lock()
	if finished[t.Name()] {
		lifecycleMu.Unlock()
		return
	}
	lifecycleMu.Unlock()

	t.Cleanup(func() {
		lifecycleMu.Lock()
		if finished[t.Name()] {
			lifecycleMu.Unlock()
			return
		}
		finished[t.Name()] = true
		lifecycleMu.Unlock()

		if t.Failed() {
			t.Logf("e2e: FAIL %s", t.Name())
		} else {
			t.Logf("e2e: PASS %s", t.Name())
		}
	})
}
