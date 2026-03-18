# pink-whisper

Whisper.cpp daemon: TCP server for speech-to-text with automatic setup. Detects hardware (CoreML/CUDA/CPU) and downloads the right binary + model.

## Install

Download binary from [Releases](https://github.com/pink-tools/pink-whisper/releases), or via pink-orchestrator:

```bash
pink-orchestrator --service-download pink-whisper
```

## Usage

```bash
pink-whisper              # Start daemon (TCP server on port 7465)
pink-whisper stop         # Stop daemon
pink-whisper status       # Check if running
pink-whisper setup        # Download whisper-server binary + model (~3GB)
pink-whisper setup --check # Check if setup is complete
```

On first run, use `pink-whisper setup` to download:
- Platform-specific whisper-server binary (CoreML on Apple Silicon, CUDA on NVIDIA, CPU fallback)
- Whisper Large V3 model from HuggingFace (~3GB)

## Protocol

Binary protocol over TCP on `127.0.0.1:7465`:

```
Request:  [4 bytes LE size][16-bit PCM, 16kHz, mono]
Response: [4 bytes LE size][UTF-8 text]
```

```python
import socket, struct
sock = socket.create_connection(("127.0.0.1", 7465))
sock.send(struct.pack("<I", len(pcm)) + pcm)
size = struct.unpack("<I", sock.recv(4))[0]
text = sock.recv(size).decode()
```

## Pre-built Server Binaries

| Platform | Acceleration |
|----------|-------------|
| macOS ARM64 | CoreML |
| Linux x64 | CUDA 12 |
| Linux x64 | CPU |
| Windows x64 | CUDA 12 |
| Windows x64 | CPU |

## Build from Source

```bash
git clone https://github.com/pink-tools/pink-whisper.git
cd pink-whisper
go build .
```
