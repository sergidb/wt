package runner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sergidb/wt/internal/config"
)

var defaultColors = []string{"green", "cyan", "yellow", "magenta", "blue", "red"}

var colorCodes = map[string]string{
	"red":     "\033[31m",
	"green":   "\033[32m",
	"yellow":  "\033[33m",
	"blue":    "\033[34m",
	"magenta": "\033[35m",
	"cyan":    "\033[36m",
	"white":   "\033[37m",
}

const colorReset = "\033[0m"

type runningService struct {
	name  string
	cmd   *exec.Cmd
	color string
}

// Run starts all given services in parallel with colored prefix output.
// It blocks until all services exit or a signal is received.
func Run(services map[string]config.Service, worktreePath string) error {
	if len(services) == 0 {
		return fmt.Errorf("no services to run")
	}

	// Compute max name length for aligned prefixes
	maxLen := 0
	for name := range services {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	// Build and start processes
	var running []runningService
	var wg sync.WaitGroup
	colorIdx := 0

	for name, svc := range services {
		color := svc.Color
		if color == "" {
			color = defaultColors[colorIdx%len(defaultColors)]
			colorIdx++
		}

		dir := filepath.Join(worktreePath, svc.Dir)

		cmd := exec.Command("sh", "-c", svc.Cmd)
		cmd.Dir = dir
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		// Merge environment
		cmd.Env = os.Environ()
		for k, v := range svc.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("stdout pipe for %s: %w", name, err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("stderr pipe for %s: %w", name, err)
		}

		prefix := formatPrefix(name, maxLen, color)

		fmt.Printf("%s Starting %s...\n", prefix, svc.Cmd)

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("starting %s: %w", name, err)
		}

		running = append(running, runningService{name: name, cmd: cmd, color: color})

		// Stream stdout and stderr with prefixes
		wg.Add(2)
		go streamLines(&wg, stdout, prefix)
		go streamLines(&wg, stderr, prefix)
	}

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either all processes to finish or a signal
	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		// All pipes closed, wait for processes
	case sig := <-sigCh:
		fmt.Printf("\n%sReceived %s, shutting down...%s\n", colorCodes["yellow"], sig, colorReset)
		shutdown(running)
	}

	// Wait for all processes to finish
	var firstErr error
	for _, rs := range running {
		if err := rs.cmd.Wait(); err != nil && firstErr == nil {
			// Ignore signal-caused exits
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() == -1 {
					continue // killed by signal
				}
			}
			firstErr = err
		}
	}

	return firstErr
}

func shutdown(services []runningService) {
	// Send SIGINT to all process groups
	for _, rs := range services {
		if rs.cmd.Process != nil {
			_ = syscall.Kill(-rs.cmd.Process.Pid, syscall.SIGINT)
		}
	}

	// Wait up to 5 seconds, then SIGKILL
	done := make(chan struct{})
	go func() {
		for _, rs := range services {
			_ = rs.cmd.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		for _, rs := range services {
			if rs.cmd.Process != nil {
				_ = syscall.Kill(-rs.cmd.Process.Pid, syscall.SIGKILL)
			}
		}
	}
}

func streamLines(wg *sync.WaitGroup, r io.Reader, prefix string) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Printf("%s %s\n", prefix, scanner.Text())
	}
}

func formatPrefix(name string, maxLen int, color string) string {
	padded := name + strings.Repeat(" ", maxLen-len(name))
	code := colorCodes[color]
	if code == "" {
		code = colorCodes["white"]
	}
	return fmt.Sprintf("%s[%s]%s", code, padded, colorReset)
}
