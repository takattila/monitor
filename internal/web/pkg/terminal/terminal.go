package terminal

import (
	"fmt"
	"os"
	"strings"

	"github.com/takattila/monitor/pkg/logger"
	"golang.org/x/crypto/ssh/terminal"
)

// Prompt executes interactive command line prompt.
func Prompt(prompt string) (text string) {
	l := logger.New(logger.ErrorLevel, logger.ColorOn)
	fmt.Print(prompt)

	str, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	l.Error(err)
	text = strings.TrimSpace(string(str))
	println()

	return strings.TrimSpace(text)
}
