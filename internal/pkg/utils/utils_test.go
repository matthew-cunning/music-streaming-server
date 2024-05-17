package utils_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matthew-c-atu/project-audio-streamer/internal/pkg/utils"
)

func TestFindFilesWithExtension(t *testing.T) {
	t.Run("test invalid filepath", func(t *testing.T) {
		invalidPath := "foo/bar"
		found := utils.FindFilesWithExtension(invalidPath, ".m3u8")
		expected := 0
		actual := len(found)
		if expected != actual {
			t.Errorf("Expected %v but got %v", expected, actual)
		}
	})
}

func TestAddHeaders(t *testing.T) {
	t.Run("test headers applied", func(t *testing.T) {
		server := httptest.NewServer(utils.AddHeaders(nil))
		defer server.Close()
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Error(err)
		}
		actual := resp.Header.Get("Cache-Control")
		expected := "no-cache, no-store"
		if expected != actual {
			t.Errorf("Expected %s but got %s", expected, actual)
		}

		fmt.Printf("expected: %s, actual: %s\n", expected, actual)
	})
}
