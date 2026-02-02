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

	cmd := exec.CommandContext(ctx, binary, model)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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
