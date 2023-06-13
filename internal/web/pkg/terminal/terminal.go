package terminal

import (
	"fmt"
	"os"
	"strings"

	"github.com/takattila/monitor/pkg/common"
	"golang.org/x/crypto/ssh/terminal"
)

// Prompt executes interactive command line prompt.
func Prompt(prompt string) (text string) {
	fmt.Print(prompt)

	str, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	common.Error(err)
	text = strings.TrimSpace(string(str))
	println()

	return strings.TrimSpace(text)
}
