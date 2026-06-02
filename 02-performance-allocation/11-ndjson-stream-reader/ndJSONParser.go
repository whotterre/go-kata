package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

func ReadNDJSON(ctx context.Context, r io.Reader, handle func([]byte) error) error {
	reader := bufio.NewReader(r)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Read a line
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				return fmt.Errorf("read error: %w", err)
			}

			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")

			if len(line) > 0 {
				if err := handle([]byte(line)); err != nil {
					return fmt.Errorf("handle error: %w", err)
				}
			}

			if err == io.EOF {
				return nil
			}
		}
	}
}

func handle(x []byte) error {
	fmt.Println(string(x))
	return nil
}
