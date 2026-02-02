package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pink-tools/pink-core"
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
	logFile := filepath.Join(dir, "whisper.log")

	// Open log file for whisper server output
	log, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}
	defer log.Close()

	cmd := exec.CommandContext(ctx, binary, model)
	cmd.Dir = dir
	cmd.Stdout = log
	cmd.Stderr = log

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start whisper server: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		cmd.Process.Kill()
		return nil
	}
}
