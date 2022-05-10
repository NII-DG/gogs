package ipfs

//mockファイルの生成
//mockgen -source ipfs_exec.go -destination mock/mock_ipfs_exec.go

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

var (
	ErrParentNotExist    = errors.New("parent does not exist")
	ErrSubmoduleNotExist = errors.New("submodule does not exist")
	ErrRevisionNotExist  = errors.New("revision does not exist")
	ErrRemoteNotExist    = errors.New("remote does not exist")
	ErrExecTimeout       = errors.New("execution was timed out")
	ErrNoMergeBase       = errors.New("no merge based was found")
	ErrNotBlob           = errors.New("the entry is not a blob")
)

const DefaultTimeout = time.Minute

const (
	CAT   = "cat"
	FILES = "files"
)

type IFIpfsCommand interface {
	AddArgs(args ...string)
	RemoveArgs()
	AddEnvs(envs ...string)
	RemoveEnvs()
	RunInDirPipelineWithTimeout(timeout time.Duration, stdout, stderr io.Writer, dir string) (err error)
	RunInDirWithTimeout(timeout time.Duration, dir string) ([]byte, error)
	RunWithTimeout(timeout time.Duration) ([]byte, error)
	RunInDir(dir string) ([]byte, error)
	Run() ([]byte, error)
}

type IpfsCommand struct {
	name string
	args []string
	envs []string
}

func NewCommand(args ...string) *IpfsCommand {
	return &IpfsCommand{
		name: "ipfs",
		args: args,
	}
}
func (c *IpfsCommand) AddArgs(args ...string) {
	c.args = append(c.args, args...)

}
func (c *IpfsCommand) RemoveArgs() {
	c.args = []string{}
}

func (c *IpfsCommand) AddEnvs(envs ...string) {
	c.envs = append(c.envs, envs...)

}

func (c *IpfsCommand) RemoveEnvs() {
	c.envs = []string{}
}

func (c *IpfsCommand) RunInDirPipelineWithTimeout(timeout time.Duration, stdout, stderr io.Writer, dir string) (err error) {
	if timeout < time.Nanosecond {
		timeout = DefaultTimeout
	}

	buf := new(bytes.Buffer)
	w := stdout
	if logOutput != nil {
		buf.Grow(512)
		w = &limitDualWriter{
			W: buf,
			N: int64(buf.Cap()),
			w: stdout,
		}
	}

	defer func() {
		if len(dir) == 0 {
			log("[timeout: %v] %s\n%s", timeout, c, buf.Bytes())
		} else {
			log("[timeout: %v] %s: %s\n%s", timeout, dir, c, buf.Bytes())
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer func() {
		cancel()
		if err == context.DeadlineExceeded {
			err = ErrExecTimeout
		}
	}()

	cmd := exec.CommandContext(ctx, c.name, c.args...)
	if len(c.envs) > 0 {
		cmd.Env = append(os.Environ(), c.envs...)
	}
	cmd.Dir = dir
	cmd.Stdout = w
	cmd.Stderr = stderr
	if err = cmd.Start(); err != nil {
		return err
	}

	result := make(chan error)
	go func() {
		result <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		<-result
		if cmd.Process != nil && cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
			if err := cmd.Process.Kill(); err != nil {
				return fmt.Errorf("kill process: %v", err)
			}
		}

		return ErrExecTimeout
	case err = <-result:
		return err
	}
}

// RunInDirWithTimeout executes the command in given directory and timeout duration.
// It returns stdout in []byte and error (combined with stderr).
func (c *IpfsCommand) RunInDirWithTimeout(timeout time.Duration, dir string) ([]byte, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if err := c.RunInDirPipelineWithTimeout(timeout, stdout, stderr, dir); err != nil {
		return nil, concatenateError(err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// RunInDir executes the command in given directory and default timeout duration.
// It returns stdout and error (combined with stderr).
func (c *IpfsCommand) RunInDir(dir string) ([]byte, error) {
	return c.RunInDirWithTimeout(DefaultTimeout, dir)
}

// RunWithTimeout executes the command in working directory and given timeout duration.
// It returns stdout in string and error (combined with stderr).
func (c *IpfsCommand) RunWithTimeout(timeout time.Duration) ([]byte, error) {
	stdout, err := c.RunInDirWithTimeout(timeout, "")
	if err != nil {
		return nil, err
	}
	return stdout, nil
}

// Run executes the command in working directory and default timeout duration.
// It returns stdout in string and error (combined with stderr).
func (c *IpfsCommand) Run() ([]byte, error) {
	return c.RunWithTimeout(DefaultTimeout)
}

func concatenateError(err error, stderr string) error {
	if len(stderr) == 0 {
		return err
	}
	return fmt.Errorf("%v - %s", err, stderr)
}

var (
	// logOutput is the writer to write logs. When not set, no log will be produced.
	logOutput io.Writer
	// logPrefix is the prefix prepend to each log entry.
	logPrefix = "[ipfs-exec] "
)

type limitDualWriter struct {
	W        io.Writer // underlying writer
	N        int64     // max bytes remaining
	prompted bool

	w io.Writer
}

func (w *limitDualWriter) Write(p []byte) (int, error) {
	if w.N > 0 {
		limit := int64(len(p))
		if limit > w.N {
			limit = w.N
		}
		n, _ := w.W.Write(p[:limit])
		w.N -= int64(n)
	}

	if !w.prompted && w.N <= 0 {
		w.prompted = true
		_, _ = w.W.Write([]byte("... (more omitted)"))
	}

	return w.w.Write(p)
}

func log(format string, args ...interface{}) {
	if logOutput == nil {
		return
	}

	fmt.Fprint(logOutput, logPrefix)
	fmt.Fprintf(logOutput, format, args...)
	fmt.Fprintln(logOutput)
}
