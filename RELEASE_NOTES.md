# Whisper.cpp Server Binaries

## Artifacts

- `darwin-arm64-coreml.tar.gz` - macOS Apple Silicon (CoreML, fast)
- `linux-amd64-cuda.tar.gz` - Linux x64 (CUDA 12, fast)
- `linux-amd64-cpu.tar.gz` - Linux x64 (CPU, slow)
- `windows-amd64-cuda.zip` - Windows x64 (CUDA 12, fast)
- `windows-amd64-cpu.zip` - Windows x64 (CPU, slow)

## Required

Model: [ggml-large-v3.bin](https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin) (~3GB)

## Structure

```
whisper-server            # binary
ggml-large-v3.bin         # model
ggml-large-v3-encoder.mlmodelc/  # macOS only
libcublas*.so / cublas*.dll      # CUDA only
```

## Note

CPU builds are slow (10-30x realtime). If no GPU, consider using the remote server at transcribe.pinkhaired.com instead.
