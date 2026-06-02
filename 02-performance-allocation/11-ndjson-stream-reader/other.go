package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

func slowParser(ctx context.Context, r io.Reader, handle func([]byte) error) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line := scanner.Text()

			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")

			if len(line) > 0 {
				if err := handle([]byte(line)); err != nil {
					return fmt.Errorf("handle error: %w", err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan error: %w", err)
	}

	return nil
}
