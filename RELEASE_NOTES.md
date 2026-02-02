# Whisper.cpp Server Binaries

GPU-accelerated only. CPU not supported — use remote server instead.

## Artifacts

- `darwin-arm64-coreml.tar.gz` - macOS Apple Silicon (CoreML)
- `linux-amd64-cuda.tar.gz` - Linux x64 (CUDA 12)
- `windows-amd64-cuda.zip` - Windows x64 (CUDA 12)

## Required

Model: [ggml-large-v3.bin](https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin) (~3GB)

## Structure

```
pink-whisper              # binary
ggml-large-v3.bin         # model
ggml-large-v3-encoder.mlmodelc/  # macOS only
libcublas*.so / cublas*.dll      # CUDA only
```
