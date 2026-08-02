package terminal

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/creack/pty"
	"github.com/stretchr/testify/require"
)

func TestDetectShellFromPasswd(t *testing.T) {
	oldUser, oldRead := currentUser, readPasswdFile
	currentUser = func() (*user.User, error) {
		return &user.User{Username: "tester", HomeDir: "/home/tester"}, nil
	}
	readPasswdFile = func() ([]byte, error) {
		return []byte("tester:x:1000:1000::/home/tester:/bin/bash\n"), nil
	}
	defer func() { currentUser, readPasswdFile = oldUser, oldRead }()

	require.Equal(t, "/bin/bash", DetectShell())
	require.Equal(t, "/home/tester", HomeDir())
}

func TestDetectShellFromShellEnv(t *testing.T) {
	oldUser, oldRead, oldGetenv := currentUser, readPasswdFile, getenv
	currentUser = func() (*user.User, error) {
		return &user.User{Username: "tester", HomeDir: ""}, nil
	}
	readPasswdFile = func() ([]byte, error) { return []byte("root:x:0:0:root:/root:/bin/bash\n"), nil }
	getenv = func(key string) string {
		if key == "SHELL" {
			return "/bin/zsh"
		}
		return ""
	}
	defer func() { currentUser, readPasswdFile, getenv = oldUser, oldRead, oldGetenv }()

	require.Equal(t, "/bin/zsh", DetectShell())
}

func TestDetectShellCandidatesBash(t *testing.T) {
	oldUser, oldRead, oldGetenv, oldFe, oldCands := currentUser, readPasswdFile, getenv, fileExists, shellCandidates
	currentUser = func() (*user.User, error) { return &user.User{Username: "tester"}, nil }
	readPasswdFile = func() ([]byte, error) { return []byte("root:x:0:0:root:/root:/bin/bash\n"), nil }
	getenv = func(string) string { return "" }
	fileExists = func(path string) bool { return path == "/bin/bash" }
	shellCandidates = []string{"/bin/bash", "/bin/sh"}
	defer func() {
		currentUser, readPasswdFile, getenv, fileExists, shellCandidates = oldUser, oldRead, oldGetenv, oldFe, oldCands
	}()

	require.Equal(t, "/bin/bash", DetectShell())
}

func TestDetectShellCandidatesSh(t *testing.T) {
	oldUser, oldRead, oldGetenv, oldFe, oldCands := currentUser, readPasswdFile, getenv, fileExists, shellCandidates
	currentUser = func() (*user.User, error) { return &user.User{Username: "tester"}, nil }
	readPasswdFile = func() ([]byte, error) { return []byte("root:x:0:0:root:/root:/bin/bash\n"), nil }
	getenv = func(string) string { return "" }
	fileExists = func(path string) bool { return path == "/bin/sh" }
	shellCandidates = []string{"/bin/bash", "/bin/sh"}
	defer func() {
		currentUser, readPasswdFile, getenv, fileExists, shellCandidates = oldUser, oldRead, oldGetenv, oldFe, oldCands
	}()

	require.Equal(t, "/bin/sh", DetectShell())
}

func TestDetectShellFallback(t *testing.T) {
	oldUser, oldRead, oldGetenv, oldFe := currentUser, readPasswdFile, getenv, fileExists
	currentUser = func() (*user.User, error) { return &user.User{Username: "tester"}, nil }
	readPasswdFile = func() ([]byte, error) { return []byte("root:x:0:0:root:/root:/bin/bash\n"), nil }
	getenv = func(string) string { return "" }
	fileExists = func(string) bool { return false }
	defer func() { currentUser, readPasswdFile, getenv, fileExists = oldUser, oldRead, oldGetenv, oldFe }()

	require.Equal(t, "sh", DetectShell())
}

func TestDetectShellUserError(t *testing.T) {
	oldUser, oldRead, oldGetenv := currentUser, readPasswdFile, getenv
	currentUser = func() (*user.User, error) { return nil, errors.New("no user") }
	readPasswdFile = func() ([]byte, error) { return []byte("root:x:0:0:root:/root:/bin/bash\n"), nil }
	getenv = func(key string) string {
		if key == "SHELL" {
			return "/bin/fish"
		}
		return ""
	}
	defer func() { currentUser, readPasswdFile, getenv = oldUser, oldRead, oldGetenv }()

	require.Equal(t, "/bin/fish", DetectShell())
}

func TestShellFromPasswdReadError(t *testing.T) {
	oldRead := readPasswdFile
	readPasswdFile = func() ([]byte, error) { return nil, errors.New("read error") }
	defer func() { readPasswdFile = oldRead }()

	require.Equal(t, "", ShellFromPasswd("tester"))
}

func TestShellFromPasswdRealFile(t *testing.T) {
	u, err := user.Current()
	require.NoError(t, err)
	require.NotEqual(t, "", ShellFromPasswd(u.Username))
}

func TestParseShellFromPasswdMatch(t *testing.T) {
	require.Equal(t, "/bin/sh", ParseShellFromPasswd("root:x:0:0:root:/root:/bin/bash\ntester:x:1000:1000::/home/tester:/bin/sh\n", "tester"))
}

func TestParseShellFromPasswdNoMatch(t *testing.T) {
	require.Equal(t, "", ParseShellFromPasswd("root:x:0:0:root:/root:/bin/bash\n", "nobody"))
}

func TestParseShellFromPasswdShortLine(t *testing.T) {
	require.Equal(t, "", ParseShellFromPasswd("short-line\n", "short-line"))
}

func TestHomeDirEnv(t *testing.T) {
	oldUser, oldGetenv := currentUser, getenv
	currentUser = func() (*user.User, error) { return nil, errors.New("no user") }
	getenv = func(key string) string {
		if key == "HOME" {
			return "/custom/home"
		}
		return ""
	}
	defer func() { currentUser, getenv = oldUser, oldGetenv }()

	require.Equal(t, "/custom/home", HomeDir())
}

func TestFileExistsReal(t *testing.T) {
	require.True(t, fileExists("/bin/sh"))
	require.False(t, fileExists("/nonexistent-file-for-monitor-test"))
}

func TestAcquireRelease(t *testing.T) {
	require.True(t, acquire())
	require.False(t, acquire())
	release()
	require.True(t, acquire())
	release()
	require.Equal(t, 0, activeSessions)
}

func TestStartSessionEmptyShellDir(t *testing.T) {
	oldUser, oldRead, oldGetenv := currentUser, readPasswdFile, getenv
	currentUser = func() (*user.User, error) {
		return &user.User{Username: "tester", HomeDir: "/tmp"}, nil
	}
	readPasswdFile = func() ([]byte, error) { return []byte("tester:x:1000:1000::/tmp:/bin/sh\n"), nil }
	getenv = func(string) string { return "" }
	defer func() { currentUser, readPasswdFile, getenv = oldUser, oldRead, oldGetenv }()

	sess, err := StartSession("", "", 0, 0)
	require.NoError(t, err)
	defer sess.Close()

	require.NotNil(t, sess.PTY)
	require.Equal(t, "/bin/sh", sess.CMD.Path)
	require.Equal(t, "/tmp", sess.CMD.Dir)

	rows, cols, err := pty.Getsize(sess.PTY)
	require.NoError(t, err)
	require.Equal(t, 24, rows)
	require.Equal(t, 80, cols)
}

func TestStartSessionError(t *testing.T) {
	_, err := StartSession("/nonexistent/shell", "", 80, 24)
	require.Error(t, err)
}

func TestSessionCloseNilFields(t *testing.T) {
	sess := &Session{}
	sess.Close()
}

func TestSessionResize(t *testing.T) {
	sess, err := StartSession("/bin/sh", "/tmp", 80, 24)
	require.NoError(t, err)
	defer sess.Close()

	sess.Resize(0, 0)

	sess.Resize(100, 40)
	rows, cols, err := pty.Getsize(sess.PTY)
	require.NoError(t, err)
	require.Equal(t, 40, rows)
	require.Equal(t, 100, cols)
}

func TestSessionWriteRead(t *testing.T) {
	sess, err := StartSession("/bin/sh", "/tmp", 80, 24)
	require.NoError(t, err)
	defer sess.Close()

	require.NoError(t, sess.Write("echo HELLO_SESSION\r"))

	deadline := time.Now().Add(5 * time.Second)
	output := ""
	for time.Now().Before(deadline) {
		data, err := sess.Read()
		if err != nil {
			break
		}
		output += string(data)
		if strings.Contains(output, "HELLO_SESSION") {
			break
		}
	}
	require.Contains(t, output, "HELLO_SESSION")
}

func TestShellEnvSetsTermWhenDumb(t *testing.T) {
	oldEnviron, oldGetenv := environ, getenv
	environ = func() []string { return []string{"PATH=/usr/bin", "HOME=/root", "TERM=dumb"} }
	getenv = func(key string) string {
		if key == "TERM" {
			return "dumb"
		}
		return ""
	}
	defer func() { environ, getenv = oldEnviron, oldGetenv }()

	require.Equal(t, []string{"PATH=/usr/bin", "HOME=/root", "TERM=xterm-256color"}, shellEnv())
}

func TestUserForEmpty(t *testing.T) {
	require.Nil(t, userFor(""))
}

func TestUserForUnknown(t *testing.T) {
	require.Nil(t, userFor("this-user-does-not-exist-xyz"))
}

func TestUserForCurrent(t *testing.T) {
	u, err := user.Current()
	require.NoError(t, err)
	got := userFor(u.Username)
	require.NotNil(t, got)
	require.Equal(t, u.Uid, got.Uid)
}

func TestShellForNilUser(t *testing.T) {
	require.Equal(t, DetectShell(), shellFor(nil))
}

func TestHomeForNilUser(t *testing.T) {
	require.Equal(t, HomeDir(), homeFor(nil))
}

func TestCredentialForNilUser(t *testing.T) {
	require.Nil(t, credentialFor(nil))
}

func TestCredentialForCurrentUser(t *testing.T) {
	u, err := user.Current()
	require.NoError(t, err)
	require.Nil(t, credentialFor(u))
}

func TestCredentialForOtherUser(t *testing.T) {
	oldGeteuid := geteuid
	geteuid = func() int { return 99999 }
	defer func() { geteuid = oldGeteuid }()

	cred := credentialFor(&user.User{Username: "root", Uid: "0", Gid: "0", HomeDir: "/root"})
	require.NotNil(t, cred)
	require.Equal(t, uint32(0), cred.Uid)
	require.Equal(t, uint32(0), cred.Gid)
	require.Contains(t, cred.Groups, uint32(0))
}

func TestCredentialForInvalidUid(t *testing.T) {
	oldGeteuid := geteuid
	geteuid = func() int { return 99999 }
	defer func() { geteuid = oldGeteuid }()

	require.Nil(t, credentialFor(&user.User{Username: "x", Uid: "not-a-number", Gid: "0", HomeDir: "/x"}))
}

func TestStartSessionAsConfiguredUser(t *testing.T) {
	oldLookup := userLookup
	cu, err := user.Current()
	require.NoError(t, err)
	userLookup = func(username string) (*user.User, error) {
		if username == "tester" {
			return cu, nil
		}
		return nil, os.ErrNotExist
	}
	defer func() { userLookup = oldLookup }()

	sess, err := StartSession("", "", 0, 0, "tester")
	require.NoError(t, err)
	defer sess.Close()

	require.Equal(t, cu.HomeDir, sess.CMD.Dir)
	require.Nil(t, sess.CMD.SysProcAttr.Credential)
	require.Contains(t, sess.CMD.Env, "HOME="+cu.HomeDir)
	require.Contains(t, sess.CMD.Env, "USER="+cu.Username)
	require.Contains(t, sess.CMD.Env, "LOGNAME="+cu.Username)
	require.Contains(t, sess.CMD.Env, "HISTFILE="+filepath.Join(cu.HomeDir, ".bash_history"))
}

func TestStartSessionWithUnknownUserFallsBack(t *testing.T) {
	oldLookup := userLookup
	userLookup = func(username string) (*user.User, error) {
		return nil, os.ErrNotExist
	}
	defer func() { userLookup = oldLookup }()

	sess, err := StartSession("", "", 0, 0, "no-such-user")
	require.NoError(t, err)
	defer sess.Close()

	require.Equal(t, HomeDir(), sess.CMD.Dir)
}

func TestStartSessionWithoutUserKeepsShell(t *testing.T) {
	oldUser, oldRead, oldGetenv := currentUser, readPasswdFile, getenv
	currentUser = func() (*user.User, error) {
		return &user.User{Username: "tester", HomeDir: "/tmp"}, nil
	}
	readPasswdFile = func() ([]byte, error) { return []byte("tester:x:1000:1000::/tmp:/bin/sh\n"), nil }
	getenv = func(string) string { return "" }
	defer func() { currentUser, readPasswdFile, getenv = oldUser, oldRead, oldGetenv }()

	sess, err := StartSession("/bin/bash", "/tmp", 80, 24)
	require.NoError(t, err)
	defer sess.Close()

	require.Equal(t, "/bin/bash", sess.CMD.Path)
	require.Equal(t, []string{"/bin/bash", "-l"}, sess.CMD.Args)
	require.Equal(t, "/tmp", sess.CMD.Dir)
	require.Nil(t, sess.CMD.SysProcAttr.Credential)
}

func TestShellEnvSetsTermWhenEmpty(t *testing.T) {
	oldEnviron, oldGetenv := environ, getenv
	environ = func() []string { return []string{"PATH=/usr/bin"} }
	getenv = func(string) string { return "" }
	defer func() { environ, getenv = oldEnviron, oldGetenv }()

	require.Equal(t, []string{"PATH=/usr/bin", "TERM=xterm-256color"}, shellEnv())
}

func TestShellEnvKeepsExistingTerm(t *testing.T) {
	oldEnviron, oldGetenv := environ, getenv
	environ = func() []string { return []string{"PATH=/usr/bin", "TERM=screen-256color"} }
	getenv = func(key string) string {
		if key == "TERM" {
			return "screen-256color"
		}
		return ""
	}
	defer func() { environ, getenv = oldEnviron, oldGetenv }()

	require.Equal(t, []string{"PATH=/usr/bin", "TERM=screen-256color"}, shellEnv())
}

func TestStartSessionBashIsLoginShell(t *testing.T) {
	sess, err := StartSession("/bin/bash", "/tmp", 80, 24)
	require.NoError(t, err)
	defer sess.Close()

	require.Equal(t, "/bin/bash", sess.CMD.Path)
	require.Equal(t, []string{"/bin/bash", "-l"}, sess.CMD.Args)
}

func TestStartSessionShNotLoginShell(t *testing.T) {
	sess, err := StartSession("/bin/sh", "/tmp", 80, 24)
	require.NoError(t, err)
	defer sess.Close()

	require.Equal(t, []string{"/bin/sh"}, sess.CMD.Args)
}

func TestSessionTermIsXterm(t *testing.T) {
	oldEnviron, oldGetenv := environ, getenv
	environ = func() []string { return []string{"PATH=/usr/bin", "HOME=/root"} }
	getenv = func(string) string { return "" }
	defer func() { environ, getenv = oldEnviron, oldGetenv }()

	sess, err := StartSession("/bin/sh", "/tmp", 80, 24)
	require.NoError(t, err)
	defer sess.Close()

	require.Equal(t, []string{"PATH=/usr/bin", "HOME=/root", "TERM=xterm-256color"}, sess.CMD.Env)

	require.NoError(t, sess.Write("echo $TERM\r"))

	deadline := time.Now().Add(5 * time.Second)
	output := ""
	for time.Now().Before(deadline) {
		data, err := sess.Read()
		if err != nil {
			break
		}
		output += string(data)
		if strings.Contains(output, "xterm-256color") {
			break
		}
	}
	require.Contains(t, output, "xterm-256color")
}

func TestShellHistoryPersistsAcrossSessions(t *testing.T) {
	hist := filepath.Join(t.TempDir(), ".bash_history")
	oldHist, oldHome := os.Getenv("HISTFILE"), os.Getenv("HOME")
	require.NoError(t, os.Setenv("HISTFILE", hist))
	t.Cleanup(func() {
		_ = os.Setenv("HISTFILE", oldHist)
		_ = os.Setenv("HOME", oldHome)
	})

	sess, err := StartSession("/bin/bash", t.TempDir(), 80, 24)
	require.NoError(t, err)

	require.NoError(t, sess.Write("echo HISTORY_PROBE_1234\r"))

	deadline := time.Now().Add(5 * time.Second)
	output := ""
	for time.Now().Before(deadline) {
		data, err := sess.Read()
		if err != nil {
			break
		}
		output += string(data)
		if strings.Contains(output, "HISTORY_PROBE_1234") {
			break
		}
	}
	require.Contains(t, output, "HISTORY_PROBE_1234")

	require.NoError(t, sess.Write("exit\r"))

	deadline = time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		_, err := sess.Read()
		if err != nil {
			break
		}
	}
	sess.Close()

	written, err := os.ReadFile(hist)
	require.NoError(t, err)
	require.Contains(t, string(written), "HISTORY_PROBE_1234")
}

func TestSessionCloseKillsStuckShell(t *testing.T) {
	script := filepath.Join(t.TempDir(), "stuck.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\ntrap '' HUP\nwhile true; do sleep 1; done\n"), 0755))

	start := time.Now()
	sess, err := StartSession(script, "/tmp", 80, 24)
	require.NoError(t, err)
	require.NotNil(t, sess.CMD.Process)

	time.Sleep(500 * time.Millisecond)
	sess.Close()

	require.Less(t, time.Since(start), 10*time.Second)
	require.NotNil(t, sess.CMD.ProcessState)
	require.Equal(t, -1, sess.CMD.ProcessState.ExitCode())
}

func newServeServer(t *testing.T, shell, dir string) (*httptest.Server, string, func()) {
	t.Helper()
	done := make(chan struct{})
	var once sync.Once
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("Accept: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")
		_ = Serve(conn, shell, dir, 80, 24)
		once.Do(func() { close(done) })
	}))
	wait := func() {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatalf("timed out waiting for Serve to finish")
		}
	}
	return srv, "ws" + strings.TrimPrefix(srv.URL, "http"), wait
}

func TestServeFlow(t *testing.T) {
	srv, url, wait := newServeServer(t, "/bin/sh", "")
	defer srv.Close()

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	require.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	require.NoError(t, conn.Write(context.Background(), websocket.MessageText, []byte(`{"type":"input","data":"echo TERMTEST\r"}`)))
	require.NoError(t, conn.Write(context.Background(), websocket.MessageText, []byte(`{"type":"resize","cols":120,"rows":40}`)))
	require.NoError(t, conn.Write(context.Background(), websocket.MessageText, []byte("this is not json")))
	require.NoError(t, conn.Write(context.Background(), websocket.MessageText, []byte(`{"type":"input","data":"echo \u00e9\u0151\u0171\r"}`)))

	deadline := time.Now().Add(5 * time.Second)
	var (
		found   bool
		utf8Msg []byte
		bin     bool
	)
	for time.Now().Before(deadline) {
		msgType, data, err := conn.Read(context.Background())
		if err != nil {
			break
		}
		if msgType == websocket.MessageBinary {
			bin = true
		}
		if strings.Contains(string(data), "TERMTEST") {
			found = true
		}
		if strings.Contains(string(data), "éőű") {
			utf8Msg = data
			break
		}
	}
	require.True(t, found)
	require.True(t, bin)
	require.Contains(t, string(utf8Msg), "éőű")

	_ = conn.Close(websocket.StatusNormalClosure, "")
	wait()
}

func TestServeBusy(t *testing.T) {
	acquire()
	defer release()

	srv, url, wait := newServeServer(t, "/bin/sh", "")
	defer srv.Close()

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	require.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, data, err := conn.Read(ctx)
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(string(data)), "busy")

	_ = conn.Close(websocket.StatusNormalClosure, "")
	wait()
}

func TestServeWriteError(t *testing.T) {
	script := filepath.Join(t.TempDir(), "burst.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\nwhile true; do echo STREAM; done\n"), 0755))

	done := make(chan struct{})
	var once sync.Once
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")
		go func() {
			time.Sleep(300 * time.Millisecond)
			_ = conn.Close(websocket.StatusNormalClosure, "")
		}()
		_ = Serve(conn, script, "/tmp", 80, 24)
		once.Do(func() { close(done) })
	}))
	defer srv.Close()

	conn, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	require.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			break
		}
	}
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("timed out waiting for Serve to finish")
	}
	time.Sleep(100 * time.Millisecond)
}

func TestServeWritePTYError(t *testing.T) {
	srv, url, wait := newServeServer(t, "/bin/sh", "")
	defer srv.Close()

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	require.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	require.NoError(t, conn.Write(context.Background(), websocket.MessageText, []byte(`{"type":"input","data":"exit\r"}`)))
	time.Sleep(300 * time.Millisecond)

	require.NoError(t, conn.Write(context.Background(), websocket.MessageText, []byte(`{"type":"input","data":"echo AFTER_EXIT\r"}`)))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		_, _, err = conn.Read(ctx)
		if err != nil {
			break
		}
	}
	require.Error(t, err)
	wait()
}

func TestServeStartError(t *testing.T) {
	srv, url, wait := newServeServer(t, "/nonexistent/shell", "")
	defer srv.Close()

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	require.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err = conn.Read(ctx)
	require.Error(t, err)
	wait()
}

func TestStartSessionWithCredential(t *testing.T) {
	oldLookup, oldGeteuid := userLookup, geteuid
	userLookup = func(username string) (*user.User, error) {
		if username == "root" {
			return &user.User{Username: "root", Uid: "0", Gid: "0"}, nil
		}
		return nil, os.ErrNotExist
	}
	geteuid = func() int { return 99999 }
	defer func() { userLookup, geteuid = oldLookup, oldGeteuid }()

	sess, err := StartSession("/bin/sh", "/tmp", 80, 24, "root")
	if err == nil {
		require.NotNil(t, sess.CMD.SysProcAttr)
		require.NotNil(t, sess.CMD.SysProcAttr.Credential)
		sess.Close()
	}
}

func TestServeWithUsername(t *testing.T) {
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")
		_ = Serve(conn, "/bin/sh", "", 80, 24, "nonexistent-user")
		close(done)
	}))
	defer srv.Close()

	conn, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	require.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	require.NoError(t, conn.Write(context.Background(), websocket.MessageText, []byte(`{"type":"input","data":"echo USERNAME_OK\r"}`)))

	deadline := time.Now().Add(5 * time.Second)
	var output string
	for time.Now().Before(deadline) {
		_, data, readErr := conn.Read(context.Background())
		if readErr != nil {
			break
		}
		output += string(data)
		if strings.Contains(output, "USERNAME_OK") {
			break
		}
	}

	require.NoError(t, conn.Close(websocket.StatusNormalClosure, ""))

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("timed out waiting for Serve to finish")
	}
}
