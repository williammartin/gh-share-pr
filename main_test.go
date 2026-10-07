package main

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSharesPullRequest(t *testing.T) {
	t.Parallel()

	// Given a pull request selected by branch in another repository
	args := []string{"feature-branch", "-R", "cli/cli"}
	pr := pullRequest{
		Title:     "Fix the bug",
		URL:       "https://github.com/cli/cli/pull/123",
		Additions: 84,
		Deletions: 8,
	}
	var stdout, stderr bytes.Buffer
	var viewedArgs []string
	var copied string
	view := func(args []string, _ io.Writer) (pullRequest, error) {
		viewedArgs = args
		return pr, nil
	}
	writeClipboard := func(text string) error {
		copied = text
		return nil
	}
	app := application{
		stdout:         &stdout,
		stderr:         &stderr,
		view:           view,
		writeClipboard: writeClipboard,
	}

	// When the pull request is shared
	err := app.run(args)

	// Then the selector and flags are forwarded, and the same Markdown is printed and copied
	require.NoError(t, err)
	assert.Equal(t, args, viewedArgs)
	assert.Equal(t, "[Fix the bug](https://github.com/cli/cli/pull/123) in cli/cli (+84/-8)\n", stdout.String())
	assert.Equal(t, "[Fix the bug](https://github.com/cli/cli/pull/123) in cli/cli (+84/-8)", copied)
	assert.Empty(t, stderr.String())
}

func TestRunWarnsWhenClipboardIsUnavailable(t *testing.T) {
	t.Parallel()

	// Given a pull request and an unavailable clipboard
	pr := pullRequest{
		Title:     "Fix the bug",
		URL:       "https://github.com/cli/cli/pull/123",
		Additions: 84,
		Deletions: 8,
	}
	var stdout, stderr bytes.Buffer
	view := func([]string, io.Writer) (pullRequest, error) {
		return pr, nil
	}
	writeClipboard := func(string) error {
		return errors.New("clipboard unavailable")
	}
	app := application{
		stdout:         &stdout,
		stderr:         &stderr,
		view:           view,
		writeClipboard: writeClipboard,
	}

	// When the pull request is shared
	err := app.run(nil)

	// Then stdout remains usable and the clipboard problem is reported separately
	require.NoError(t, err)
	assert.Equal(t, "[Fix the bug](https://github.com/cli/cli/pull/123) in cli/cli (+84/-8)\n", stdout.String())
	assert.Contains(t, stderr.String(), "warning: failed to copy to clipboard: clipboard unavailable")
}

func TestRunStopsWhenPullRequestLookupFails(t *testing.T) {
	t.Parallel()

	// Given a pull request selector that cannot be resolved
	var stdout, stderr bytes.Buffer
	view := func([]string, io.Writer) (pullRequest, error) {
		return pullRequest{}, errors.New("no pull request found")
	}
	clipboardWritten := false
	writeClipboard := func(string) error {
		clipboardWritten = true
		return nil
	}
	app := application{
		stdout:         &stdout,
		stderr:         &stderr,
		view:           view,
		writeClipboard: writeClipboard,
	}

	// When the pull request is shared
	err := app.run(nil)

	// Then the lookup error is returned without printing or copying misleading output
	require.Error(t, err)
	assert.ErrorContains(t, err, "gh pr view failed: no pull request found")
	assert.Empty(t, stdout.String())
	assert.False(t, clipboardWritten, "expected the clipboard to remain unchanged after lookup failure")
}
