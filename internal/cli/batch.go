package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// batchCommand is a single command in a batch request.
type batchCommand struct {
	Cmd  string   `json:"cmd"`
	Args []string `json:"args,omitempty"`
}

// batchResult is the result of a single command in a batch.
type batchResult struct {
	Cmd   string `json:"cmd"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func (c *CLI) batchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "batch",
		Short: "Execute commands from stdin (one JSON object per line)",
		Long: `Read newline-delimited JSON commands from stdin and execute each one.
Each line is a JSON object: {"cmd": "add", "args": ["task title", "--priority", "high"]}
Output is a JSON array of results.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			scanner := bufio.NewScanner(os.Stdin)
			var results []batchResult

			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				var bc batchCommand
				if err := json.Unmarshal([]byte(line), &bc); err != nil {
					results = append(results, batchResult{
						Cmd:   line,
						OK:    false,
						Error: fmt.Sprintf("invalid JSON: %v", err),
					})
					continue
				}

				if bc.Cmd == "batch" {
					results = append(results, batchResult{Cmd: bc.Cmd, OK: false, Error: "batch cannot be nested"})
					continue
				}

				// Build a fresh command tree for each command to avoid
				// cobra flag state leaking between invocations.
				fresh := New(c.taskStore, c.journalStore, c.focusStore, c.cfg)
				fresh.SetOut(io.Discard)
				fresh.SetErr(io.Discard)
				fresh.SetArgs(append([]string{bc.Cmd}, bc.Args...))
				err := fresh.Execute()

				br := batchResult{Cmd: bc.Cmd, OK: err == nil}
				if err != nil {
					br.Error = err.Error()
				}
				results = append(results, br)
			}

			if err := scanner.Err(); err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(results)
		},
	}
}
