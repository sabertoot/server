package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func File(
	ctx context.Context,
	url string,
	path string,
) error {
	client := http.DefaultClient

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading file: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	defer file.Close()
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	return nil
}
