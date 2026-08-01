package terminal

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/coder/websocket"
	"github.com/creack/pty"
)

// Message is a client -> server WebSocket message.
type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// Session holds the shell process and its PTY.
type Session struct {
	CMD *exec.Cmd
	PTY *os.File
}

var (
	currentUser    = user.Current
	readPasswdFile = func() ([]byte, error) { return os.ReadFile("/etc/passwd") }
	fileExists     = func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}
	getenv          = os.Getenv
	environ         = os.Environ
	geteuid         = os.Geteuid
	userLookup      = user.Lookup
	shellCandidates = []string{"/bin/bash", "/bin/sh"}
)

// MaxSessions limits the number of concurrent terminal sessions.
const MaxSessions = 1

// ErrBusy is returned when a terminal session is already active.
var ErrBusy = errors.New("terminal: another session is active")

var (
	sessionMu      sync.Mutex
	activeSessions int
)

func acquire() bool {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	if activeSessions >= MaxSessions {
		return false
	}
	activeSessions++
	return true
}

func release() {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	activeSessions--
}

// DetectShell returns the login shell of the current user.
func DetectShell() string {
	if u, err := currentUser(); err == nil && u.Username != "" {
		if shell := ShellFromPasswd(u.Username); shell != "" {
			return shell
		}
	}
	if shell := getenv("SHELL"); shell != "" {
		return shell
	}
	for _, candidate := range shellCandidates {
		if fileExists(candidate) {
			return candidate
		}
	}
	return "sh"
}

// ShellFromPasswd reads the /etc/passwd file and returns the shell of the given user.
func ShellFromPasswd(username string) string {
	data, err := readPasswdFile()
	if err != nil {
		return ""
	}
	return ParseShellFromPasswd(string(data), username)
}

// ParseShellFromPasswd returns the shell of the given user from the passwd contents.
func ParseShellFromPasswd(contents, username string) string {
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) == 7 && fields[0] == username {
			return fields[6]
		}
	}
	return ""
}

// HomeDir returns the home directory of the current user.
func HomeDir() string {
	if u, err := currentUser(); err == nil && u.HomeDir != "" {
		return u.HomeDir
	}
	return getenv("HOME")
}

// userFor returns the system user entry for the given username. It returns
// nil when the username is empty or does not map to a system user.
func userFor(username string) *user.User {
	if username == "" {
		return nil
	}
	u, err := userLookup(username)
	if err != nil {
		return nil
	}
	return u
}

// shellFor returns the login shell of the given user, falling back to the
// default shell detection when it cannot be determined.
func shellFor(u *user.User) string {
	if u != nil && u.Username != "" {
		if shell := ShellFromPasswd(u.Username); shell != "" {
			return shell
		}
	}
	return DetectShell()
}

// homeFor returns the home directory of the given user, falling back to the
// default home detection when it cannot be determined.
func homeFor(u *user.User) string {
	if u != nil && u.HomeDir != "" {
		return u.HomeDir
	}
	return HomeDir()
}

// credentialFor returns the process credential used to run the shell as the
// given user. It returns nil when no identity change is needed.
func credentialFor(u *user.User) *syscall.Credential {
	if u == nil {
		return nil
	}
	uid, errUID := strconv.ParseUint(u.Uid, 10, 32)
	gid, errGID := strconv.ParseUint(u.Gid, 10, 32)
	if errUID != nil || errGID != nil {
		return nil
	}
	if uint32(uid) == uint32(geteuid()) {
		return nil
	}
	groups := make([]uint32, 0, 1)
	if groupIDs, err := u.GroupIds(); err == nil {
		for _, id := range groupIDs {
			if gidVal, err := strconv.ParseUint(id, 10, 32); err == nil {
				groups = append(groups, uint32(gidVal))
			}
		}
	}
	return &syscall.Credential{
		Uid:    uint32(uid),
		Gid:    uint32(gid),
		Groups: groups,
	}
}

// shellEnv returns the environment for the shell process with a usable
// TERM value so that line editing and tab completion work over the PTY.
// When a user is given, the environment is adjusted to that user's
// identity so that the shell reads and writes the right history file.
func shellEnv(users ...*user.User) []string {
	env := environ()
	term := getenv("TERM")
	if term == "" || term == "dumb" {
		env = removeEnv(env, "TERM")
		env = append(env, "TERM=xterm-256color")
	}
	if len(users) > 0 && users[0] != nil {
		u := users[0]
		env = setEnv(env, "HOME", u.HomeDir)
		env = setEnv(env, "USER", u.Username)
		env = setEnv(env, "LOGNAME", u.Username)
		env = setEnv(env, "SHELL", shellFor(u))
		env = setEnv(env, "HISTFILE", filepath.Join(u.HomeDir, ".bash_history"))
	}
	return env
}

// setEnv sets the given key/value pair in the environment, replacing any
// previous entry with the same key.
func setEnv(env []string, key, value string) []string {
	env = removeEnv(env, key)
	return append(env, key+"="+value)
}

// removeEnv returns the given environment without the entry matching key.
func removeEnv(env []string, key string) []string {
	out := env[:0]
	prefix := key + "="
	for _, e := range env {
		if !strings.HasPrefix(e, prefix) {
			out = append(out, e)
		}
	}
	return out
}

// StartSession starts a shell in the given directory on a new PTY. When a
// username is given and it maps to a system user, the shell is started as
// that user with the user's home directory and shell.
func StartSession(shell, dir string, cols, rows int, usernames ...string) (*Session, error) {
	var u *user.User
	if len(usernames) > 0 {
		u = userFor(usernames[0])
	}
	if u != nil {
		shell = shellFor(u)
		dir = homeFor(u)
	}
	if shell == "" {
		shell = DetectShell()
	}
	if dir == "" {
		dir = HomeDir()
	}
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	cmd := exec.Command(shell)
	if strings.HasSuffix(shell, "bash") {
		cmd.Args = []string{shell, "-l"}
	}
	cmd.Dir = dir
	cmd.Env = shellEnv(u)
	if cred := credentialFor(u); cred != nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{Credential: cred}
	}
	f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	if err != nil {
		return nil, err
	}
	return &Session{CMD: cmd, PTY: f}, nil
}

// Close kills the shell process and closes the PTY. Closing the PTY master
// sends SIGHUP to the shell (session leader), which lets bash flush its
// history to HISTFILE before it exits. The SIGKILL is only a fallback if the
// shell does not exit on its own within the grace period.
func (s *Session) Close() {
	if s.PTY != nil {
		_ = s.PTY.Close()
	}
	if s.CMD != nil && s.CMD.Process != nil {
		done := make(chan struct{})
		go func() {
			_ = s.CMD.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(750 * time.Millisecond):
			_ = s.CMD.Process.Kill()
			<-done
		}
	}
}

// Resize sets the terminal size of the PTY.
func (s *Session) Resize(cols, rows int) {
	if cols > 0 && rows > 0 {
		_ = pty.Setsize(s.PTY, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	}
}

// Write writes the given data to the PTY.
func (s *Session) Write(data string) error {
	_, err := s.PTY.Write([]byte(data))
	return err
}

// Read reads from the PTY.
func (s *Session) Read() ([]byte, error) {
	buf := make([]byte, 4096)
	n, err := s.PTY.Read(buf)
	if n > 0 {
		return buf[:n], nil
	}
	return nil, err
}

// Serve runs a shell session over the given WebSocket connection. When a
// username is given and it maps to a system user, the shell is started as
// that user.
func Serve(conn *websocket.Conn, shell, dir string, cols, rows int, usernames ...string) error {
	if !acquire() {
		_ = conn.Write(context.Background(), websocket.MessageText, []byte("\r\n\x1b[31mTerminal is busy: another session is active. Close it and retry.\x1b[0m\r\n"))
		return ErrBusy
	}
	defer release()

	var username string
	if len(usernames) > 0 {
		username = usernames[0]
	}
	session, err := StartSession(shell, dir, cols, rows, username)
	if err != nil {
		return err
	}
	defer session.Close()

	ctx := context.Background()
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			data, readErr := session.Read()
			if len(data) > 0 {
				if writeErr := conn.Write(ctx, websocket.MessageBinary, data); writeErr != nil {
					return
				}
			}
			if readErr != nil {
				_ = session.PTY.Close()
				return
			}
		}
	}()

	for {
		_, data, readErr := conn.Read(ctx)
		if readErr != nil {
			return readErr
		}
		var msg Message
		if jsonErr := json.Unmarshal(data, &msg); jsonErr != nil {
			continue
		}
		switch msg.Type {
		case "input":
			if err := session.Write(msg.Data); err != nil {
				return err
			}
		case "resize":
			session.Resize(msg.Cols, msg.Rows)
		}
	}
}
