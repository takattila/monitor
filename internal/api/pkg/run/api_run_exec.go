package run

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/takattila/monitor/pkg/common"
)

var (
	osCreate = func(stdout string) (*os.File, error) {
		return os.Create(stdout)
	}
	cmdStdoutPipe = func(cmd *exec.Cmd) (io.ReadCloser, error) {
		return cmd.StdoutPipe()
	}
	cmdStart = func(cmd *exec.Cmd) error {
		return cmd.Start()
	}
)

// Exec starts a spacific command by its name.
func Exec(name string) (err error) {
	cmd := GetRunByName(name)
	err = Run(name, cmd)
	return err
}

// GetRunByName returns a command by its name.
func GetRunByName(name string) string {
	var command string

	for cmd, _ := range Cfg.Data.GetStringMapString("on_runtime.run") {
		if cmd == name {
			slice := Cfg.Data.GetStringSlice("on_runtime.run." + cmd)
			command = strings.Join(slice, ` `)
		}
	}

	return command
}

// Run issues a specific command.
func Run(name, command string) (err error) {
	// time.Sleep(500 * time.Millisecond)

	stdout := CmdFolder + name + ".stdout"
	finish := CmdFolder + name + ".finish"

	execute := func(stdout, finish, command string) (err error) {
		cmd := exec.Command("bash", "-c", command)

		outfile, err := osCreate(stdout)
		if err != nil {
			return err
		}
		defer outfile.Close()

		stdoutPipe, err := cmdStdoutPipe(cmd)
		if err != nil {
			return err
		}

		writer := bufio.NewWriter(outfile)
		defer func() {
			_, _ = writer.WriteString("~x~o(f)o~x~")
			writer.Flush()
			// _ = os.Rename(stdout, finish)
		}()

		err = cmdStart(cmd)
		if err != nil {
			return err
		}

		go io.Copy(writer, stdoutPipe)
		cmd.Wait()

		return nil
	}

	// stdout exists and finish NOT exists
	if common.FileExists(stdout) && !common.FileExists(finish) {
		return fmt.Errorf("the command: '%s' is running  already", command)
	}
	// stdout exists and finish exists
	if common.FileExists(stdout) && common.FileExists(finish) {
		_ = os.Remove(stdout)
		_ = os.Remove(finish)
		err = execute(stdout, finish, command)
	}
	// stdout NOT exists and finish exists
	if !common.FileExists(stdout) && common.FileExists(finish) {
		_ = os.Remove(finish)
		err = execute(stdout, finish, command)
	}
	// stdout NOT exists and finish NOT exists
	if !common.FileExists(finish) && !common.FileExists(stdout) {
		err = execute(stdout, finish, command)
	}

	return err
}
