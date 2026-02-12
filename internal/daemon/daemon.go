package daemon

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pink-tools/pink-core"
	"github.com/pink-tools/pink-core/log"
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
	cmd.Stderr = io.Discard
	setProcessGroup(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start whisper server: %w", err)
	}

	// Wait for port to be open
	addr := "127.0.0.1:7465"
	for i := 0; i < 120; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	log.Info(ctx, "whisper ready", log.Attr{"port", "7465"})

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
