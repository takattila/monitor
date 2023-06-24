package auth

import (
	"bufio"
	"encoding/base64"
	"os"

	"github.com/takattila/monitor/pkg/common"
)

// Authenticate checks whether credentials exists in AuthFile or not.
func Authenticate(authFile, name, pass string) bool {
	file, err := os.Open(authFile)
	if err != nil {
		common.Error(err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	authString := base64.StdEncoding.EncodeToString([]byte(name + ":" + pass))

	for scanner.Scan() {
		if authString == scanner.Text() {
			return true
		}
	}

	return false
}
