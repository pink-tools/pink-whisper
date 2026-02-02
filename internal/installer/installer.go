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

	dir := core.DataDir("pink-whisper")
	binaryPath := filepath.Join(dir, ServerBinaryName())
	modelPath := filepath.Join(dir, "ggml-large-v3.bin")

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

func Install(info InstallInfo) error {
	dir := core.DataDir("pink-whisper")

	fmt.Printf("Downloading %s...\n", info.Artifact)
	artifactPath := filepath.Join(dir, info.Artifact)
	if err := download(releaseURL+info.Artifact, artifactPath); err != nil {
		return fmt.Errorf("download artifact: %w", err)
	}

	fmt.Println("Extracting...")
	if err := extract(artifactPath, dir); err != nil {
		return fmt.Errorf("extract: %w", err)
	}
	os.Remove(artifactPath)

	// Rename extracted binary to whisper-server (avoid conflict with Go wrapper)
	oldName := filepath.Join(dir, cppBinaryName())
	newName := filepath.Join(dir, ServerBinaryName())
	if fileExists(oldName) && oldName != newName {
		os.Rename(oldName, newName)
	}

	modelPath := filepath.Join(dir, "ggml-large-v3.bin")
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

// cppBinaryName returns the original C++ binary name from the archive
func cppBinaryName() string {
	if runtime.GOOS == "windows" {
		return "pink-whisper.exe"
	}
	return "pink-whisper"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func download(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
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
