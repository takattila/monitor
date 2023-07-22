package run

import (
	"os"

	"github.com/takattila/monitor/pkg/common"
)

// StdOut provides a specific command output.
func StdOut(name string) (content string) {
	stdout := "./cmd/" + name + ".stdout"
	finish := "./cmd/" + name + ".finish"

	if common.FileExists(stdout) {
		if b, err := os.ReadFile(stdout); err == nil {
			content = string(b)
		}
	}

	if common.FileExists(finish) {
		if b, err := os.ReadFile(finish); err == nil {
			content = string(b)
		}
	}

	return string(content)
}
