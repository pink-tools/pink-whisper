package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pink-tools/pink-core"
	"github.com/pink-tools/pink-whisper/internal/hardware"
)

const (
	releaseURL = "https://github.com/pink-tools/pink-whisper/releases/download/cpp-latest/"
	modelURL   = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin"
	modelSize  = 3095033483 // ~3GB
)

type Dialog struct {
	Title         string `json:"title"`
	Message       string `json:"message"`
	ConfirmButton string `json:"confirm_button"`
	CancelButton  string `json:"cancel_button"`
}

type InstallInfo struct {
	Ready       bool    `json:"ready"`
	NeedConfirm bool    `json:"need_confirm"`
	Dialog      *Dialog `json:"dialog,omitempty"`
	Hardware    string  `json:"hardware"`
	Artifact    string  `json:"artifact"`
	ModelSize   int64   `json:"model_size"`
}

func Check() InstallInfo {
	hw := hardware.Get()
	info := InstallInfo{
		Hardware:  string(hw),
		Artifact:  hardware.ArtifactName(),
		ModelSize: modelSize,
	}

	binaryPath := filepath.Join(core.ServiceDir("pink-whisper"), ServerBinaryName())
	modelPath := ModelPath()

	// Migrate model from old AppDataDir location
	if !fileExists(modelPath) {
		oldPath := filepath.Join(core.AppDataDir("pink-whisper"), "ggml-large-v3.bin")
		if fileExists(oldPath) {
			os.Rename(oldPath, modelPath)
		}
	}

	if fileExists(binaryPath) && fileExists(modelPath) {
		info.Ready = true
		return info
	}

	// Show confirmation dialog for CPU (slow) hardware
	if !hardware.IsFast() {
		info.NeedConfirm = true
		info.Dialog = &Dialog{
			Title: "pink-whisper",
			Message: fmt.Sprintf(`No GPU acceleration detected.

Hardware: %s

LOCAL SERVER (CPU):
- Very slow transcription (10-30x realtime)
- Downloads ~3GB model
- Uses significant CPU/RAM

REMOTE SERVER (transcribe.pinkhaired.com):
- Fast transcription (GPU accelerated)
- No local installation needed
- Works automatically with pink-transcriber`, hardware.Description()),
			ConfirmButton: "Install locally",
			CancelButton:  "Use remote (recommended)",
		}
	}

	return info
}

// ModelPath returns the path to the whisper model file.
func ModelPath() string {
	return filepath.Join(core.ServiceDir("pink-whisper"), "ggml-large-v3.bin")
}

func Install(info InstallInfo) error {
	binDir := core.ServiceDir("pink-whisper")

	fmt.Printf("Downloading %s...\n", info.Artifact)
	artifactPath := filepath.Join(binDir, info.Artifact)
	if err := download(releaseURL+info.Artifact, artifactPath); err != nil {
		return fmt.Errorf("download artifact: %w", err)
	}

	fmt.Println("Extracting...")
	if err := extract(artifactPath, binDir); err != nil {
		return fmt.Errorf("extract: %w", err)
	}
	os.Remove(artifactPath)

	// Ensure server binary is executable (tar may not preserve permissions)
	os.Chmod(filepath.Join(binDir, ServerBinaryName()), 0755)

	modelPath := ModelPath()
	if !fileExists(modelPath) {
		fmt.Printf("Downloading model (~3GB)...\n")
		if err := download(modelURL, modelPath); err != nil {
			return fmt.Errorf("download model: %w", err)
		}
	}

	fmt.Println("Installed successfully")
	return nil
}

// ServerBinaryName returns the name of the whisper server binary
func ServerBinaryName() string {
	if runtime.GOOS == "windows" {
		return "whisper-server.exe"
	}
	return "whisper-server"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func download(url, dest string) error {
	tmpFile := dest + ".tmp"

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d %s (url: %s)", resp.StatusCode, http.StatusText(resp.StatusCode), url)
	}

	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}

	total := resp.ContentLength
	var downloaded int64
	var lastPct int
	buf := make([]byte, 32*1024)

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				out.Close()
				os.Remove(tmpFile)
				return writeErr
			}
			downloaded += int64(n)
			if total > 0 {
				pct := int(float64(downloaded) / float64(total) * 100)
				if pct >= lastPct+5 || pct == 100 {
					fmt.Printf("%d%% (%s / %s)\n", pct, formatBytes(downloaded), formatBytes(total))
					lastPct = pct
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			out.Close()
			os.Remove(tmpFile)
			return readErr
		}
	}

	out.Close()

	if total > 0 && downloaded != total {
		os.Remove(tmpFile)
		return fmt.Errorf("incomplete download: got %d bytes, expected %d", downloaded, total)
	}

	if err := os.Rename(tmpFile, dest); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to finalize download: %w", err)
	}

	return nil
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func extract(archive, dest string) error {
	if strings.HasSuffix(archive, ".zip") {
		return extractZip(archive, dest)
	}
	return extractTarGz(archive, dest)
}

func extractTarGz(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(dest, hdr.Name)
		if hdr.Typeflag == tar.TypeDir {
			os.MkdirAll(path, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		out.Close()
	}
	return nil
}

func extractZip(archive, dest string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
