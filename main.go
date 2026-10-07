package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
)

const pullRequestFields = "title,url,additions,deletions"

type pullRequest struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

type application struct {
	stdout         io.Writer
	stderr         io.Writer
	view           func([]string, io.Writer) (pullRequest, error)
	writeClipboard func(string) error
}

func main() {
	app := application{
		stdout:         os.Stdout,
		stderr:         os.Stderr,
		view:           viewPullRequest,
		writeClipboard: clipboard.WriteAll,
	}

	if err := app.run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (app application) run(args []string) error {
	pr, err := app.view(args, app.stderr)
	if err != nil {
		return fmt.Errorf("gh pr view failed: %w", err)
	}

	output, err := formatPullRequest(pr)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(app.stdout, output); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	if err := app.writeClipboard(output); err != nil {
		if _, writeErr := fmt.Fprintf(app.stderr, "warning: failed to copy to clipboard: %v\n", err); writeErr != nil {
			return fmt.Errorf("write clipboard warning: %w", writeErr)
		}
	}

	return nil
}

func viewPullRequest(args []string, stderr io.Writer) (pullRequest, error) {
	commandArgs := []string{"pr", "view"}
	commandArgs = append(commandArgs, args...)
	commandArgs = append(commandArgs, "--json", pullRequestFields)

	cmd := exec.Command("gh", commandArgs...)
	cmd.Stderr = stderr

	output, err := cmd.Output()
	if err != nil {
		return pullRequest{}, err
	}

	var pr pullRequest
	if err := json.Unmarshal(output, &pr); err != nil {
		return pullRequest{}, fmt.Errorf("decode gh output: %w", err)
	}

	return pr, nil
}

func formatPullRequest(pr pullRequest) (string, error) {
	repository, err := repositoryFromURL(pr.URL)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"[%s](%s) in %s (+%d/-%d)",
		pr.Title,
		pr.URL,
		repository,
		pr.Additions,
		pr.Deletions,
	), nil
}

func repositoryFromURL(prURL string) (string, error) {
	parsedURL, err := url.Parse(prURL)
	if err != nil {
		return "", fmt.Errorf("parse pull request URL: %w", err)
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "pull" || parts[0] == "" || parts[1] == "" {
		return "", errors.New("pull request URL does not contain an owner and repository")
	}

	return parts[0] + "/" + parts[1], nil
}
