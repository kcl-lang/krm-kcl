package source

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/hashicorp/go-getter"
)

// ReadThroughGetter is used to get the source code content from different types of sources (local path, remote URL, or Git source).
// It gets the pwd, creates temp files, and builds the client to get the code content.
// If an error occurs during the acquisition process, an error message will be returned.
func ReadThroughGetter(src string, opts ...getter.ClientOption) (string, error) {
	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		return src, err
	}
	// Create temp files.
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %v", err)
	}
	// Getter context
	ctx, cancel := context.WithCancel(context.Background())
	// Build the client
	client := &getter.Client{
		Ctx:     ctx,
		Src:     src,
		Dst:     tmpDir,
		Pwd:     pwd,
		Mode:    getter.ClientModeAny,
		Options: opts,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := client.Get(); err != nil {
			errChan <- err
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	select {
	case <-c:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
	case <-ctx.Done():
		wg.Wait()
	case err := <-errChan:
		wg.Wait()
		return src, err
	}
	// Read source from the temp directory
	return tmpDir, nil
}
