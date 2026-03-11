package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pink-tools/pink-core"
	"github.com/pink-tools/pink-whisper/internal/daemon"
	"github.com/pink-tools/pink-whisper/internal/hardware"
	"github.com/pink-tools/pink-whisper/internal/installer"
)

//go:embed context.md
var claudeContext string

var version = "dev"

func main() {
	cfg := core.Config{
		Name:    "pink-whisper",
		Version: version,
		Context: claudeContext,
		Commands: map[string]core.Command{
			"stop": {
				Desc: "Stop whisper server",
				Run:  cmdStop,
			},
			"status": {
				Desc: "Check server status",
				Run:  cmdStatus,
			},
			"install": {
				Desc: "Install whisper binary and model",
				Run:  cmdInstall,
			},
		},
	}
	core.HandleActions(&cfg, nil, nil)
	core.Run(cfg, daemon.Run)
}

func cmdStop(args []string) error {
	return core.SendStop("pink-whisper")
}

func cmdStatus(args []string) error {
	if core.IsRunning("pink-whisper") {
		fmt.Println("running")
	} else {
		fmt.Println("stopped")
	}
	return nil
}

func cmdInstall(args []string) error {
	if hasFlag(args, "--describe") {
		return nil
	}

	check := hasFlag(args, "--check")
	yes := hasFlag(args, "--yes")

	info := installer.Check()

	if check {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}

	if info.Ready {
		fmt.Println("already installed")
		return nil
	}

	if info.NeedConfirm && !yes {
		if !confirmInstall(info) {
			fmt.Println("skipped")
			return nil
		}
	}

	return installer.Install(info)
}

func confirmInstall(info installer.InstallInfo) bool {
	if info.Dialog == nil {
		return true
	}

	fmt.Println()
	fmt.Println(info.Dialog.Message)
	fmt.Println()

	if info.Dialog.ConfirmButton == "" {
		fmt.Printf("[%s]\n", info.Dialog.CancelButton)
		return false
	}

	fmt.Printf("[1] %s\n", info.Dialog.ConfirmButton)
	fmt.Printf("[2] %s\n", info.Dialog.CancelButton)
	fmt.Print("\nChoice [1/2]: ")

	var choice string
	fmt.Scanln(&choice)
	return choice == "1" || choice == ""
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func init() {
	hardware.Init()
}
