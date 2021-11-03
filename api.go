package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cli/safeexec"
)

func searchRepos(opts *SearchOptions) ([]Repository, error) {
	var result []Repository
	query := fmt.Sprintf("q=%s", opts.Query)

	if opts.Topic != "" {
		query += fmt.Sprintf("+topic:%s", opts.Topic)
	}
	if opts.SearchIn != "name" {
		query += fmt.Sprintf("+in:%s", opts.SearchIn)
	}

	args := []string{
		"api", "-X", "GET", "https://api.github.com/search/repositories",
		"--jq", ".items",
		"--cache=5m",
		"-f", query,
		"-f", fmt.Sprintf("per_page=%d", opts.Limit)}

	if opts.SortBy != "" {
		args = append(args, "-f")
		args = append(args, fmt.Sprintf("sort=%s", strings.ToLower(opts.SortBy)))
	}

	stdOut, _, err := gh(args...)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(stdOut.Bytes(), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// call gh and write output to buffer
func gh(args ...string) (stdOut, errOut bytes.Buffer, err error) {
	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		err = fmt.Errorf("gh not found. error: %w", err)
		return
	}

	cmd := exec.Command(ghBin, args...)
	cmd.Stderr = &errOut
	cmd.Stdout = &stdOut

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to run gh. error: %w, stderr: %s", err, errOut.String())
		return
	}

	return
}
