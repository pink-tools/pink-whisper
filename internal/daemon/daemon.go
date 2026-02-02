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
	"github.com/pink-tools/pink-otel"
	"github.com/pink-tools/pink-whisper/internal/installer"
)

func Run(ctx context.Context) error {
	info := installer.Check()
	if !info.Ready {
		return fmt.Errorf("not installed, run: pink-whisper install")
	}

	dir := core.DataDir("pink-whisper")
	binary := filepath.Join(dir, installer.ServerBinaryName())
	model := filepath.Join(dir, "ggml-large-v3.bin")

	cmd := exec.Command(binary, model)
	cmd.Dir = dir
	cmd.Stdout = io.Discard
	setProcessGroup(cmd)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start whisper server: %w", err)
	}

	// Wait for "listening on port" message, then discard rest
	ready := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "listening on port") {
				close(ready)
				io.Copy(io.Discard, stderr)
				return
			}
		}
	}()

	<-ready
	otel.Info(ctx, "whisper ready", otel.Attr{"port", "7465"})

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
		otel.Info(ctx, "stopping whisper")
		gracefulKill(cmd)
		return nil
	}
}
