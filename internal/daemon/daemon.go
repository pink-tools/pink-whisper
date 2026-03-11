package daemon

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pink-tools/pink-core"
	"github.com/pink-tools/pink-core/log"
	"github.com/pink-tools/pink-whisper/internal/setup"
)

func Run(ctx context.Context) error {
	info := setup.Check()
	if !info.Ready {
		return fmt.Errorf("not set up, run: pink-whisper setup")
	}

	binDir := core.ServiceDir("pink-whisper")
	binary := filepath.Join(binDir, setup.ServerBinaryName())

	killStaleServer()

	cmd := exec.Command(binary, "-m", setup.ModelPath())
	cmd.Dir = binDir
	cmd.Stdout = io.Discard
	setProcessGroup(cmd)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start whisper server: %w", err)
	}

	// Wait for "listening on" message from whisper-server stderr
	ready := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "listening on") {
				close(ready)
				// Drain remaining stderr so the pipe doesn't block
				io.Copy(io.Discard, stderr)
				return
			}
		}
	}()

	select {
	case <-ready:
	case <-ctx.Done():
		gracefulKill(cmd)
		return nil
	}

	log.Info(ctx, "whisper ready", log.Attr{K: "port", V: "7465"})

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("whisper exited: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Info(ctx, "stopping whisper")
		gracefulKill(cmd)
		return nil
	}
}
