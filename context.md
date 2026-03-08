# pink-whisper

TCP server for whisper.cpp (Whisper Large V3). Provides local speech-to-text on port 7465.

    pink-whisper                    Start daemon
    pink-whisper stop               Stop daemon
    pink-whisper status             Check server status
    pink-whisper install [--check]  Install binary and model (~3GB)

Hardware detection: CoreML (Apple Silicon), CUDA (NVIDIA), CPU fallback.
Model stored in platform data directory (~/Library/Application Support/pink-tools/pink-whisper/).
