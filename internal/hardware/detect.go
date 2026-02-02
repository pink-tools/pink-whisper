package hardware

import (
	"os/exec"
	"runtime"
)

type Type string

const (
	CoreML      Type = "coreml"
	CUDA        Type = "cuda"
	Unsupported Type = "unsupported"
)

var detected Type

func Init() {
	detected = detect()
}

func Get() Type {
	return detected
}

func IsSupported() bool {
	return detected != Unsupported
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

	// Everything else → not supported (use remote server)
	return Unsupported
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
		return ""
	}
}

func Description() string {
	switch detected {
	case CoreML:
		return "Apple Silicon (CoreML)"
	case CUDA:
		return "NVIDIA GPU (CUDA)"
	default:
		return "no GPU acceleration"
	}
}
