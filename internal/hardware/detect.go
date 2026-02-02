package hardware

import (
	"os/exec"
	"runtime"
)

type Type string

const (
	CoreML Type = "coreml"
	CUDA   Type = "cuda"
	CPU    Type = "cpu"
)

var detected Type

func Init() {
	detected = detect()
}

func Get() Type {
	return detected
}

func IsFast() bool {
	return detected == CoreML || detected == CUDA
}

func detect() Type {
	// Apple Silicon Mac → CoreML
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		return CoreML
	}

	// NVIDIA GPU → CUDA
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		return CUDA
	}

	// Windows/Linux → CPU (slow, but supported)
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		return CPU
	}

	// Intel Mac and other platforms → CPU
	return CPU
}

func ArtifactName() string {
	switch detected {
	case CoreML:
		return "darwin-arm64-coreml.tar.gz"
	case CUDA:
		if runtime.GOOS == "windows" {
			return "windows-amd64-cuda.zip"
		}
		return "linux-amd64-cuda.tar.gz"
	default:
		if runtime.GOOS == "windows" {
			return "windows-amd64-cpu.zip"
		}
		return "linux-amd64-cpu.tar.gz"
	}
}

func Description() string {
	switch detected {
	case CoreML:
		return "Apple Silicon (CoreML)"
	case CUDA:
		return "NVIDIA GPU (CUDA)"
	default:
		return "CPU (no GPU acceleration)"
	}
}
