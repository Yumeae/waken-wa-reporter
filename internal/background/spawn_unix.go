//go:build !windows

package background

import (
	"os"
	"os/exec"
)

// SpawnDetached starts the same executable without -background, with WAKEN_BACKGROUND_CHILD=1.
func SpawnDetached(argv []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := stripBackgroundFlag(argv)
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), "WAKEN_BACKGROUND_CHILD=1")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func stripBackgroundFlag(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "-background" {
			continue
		}
		out = append(out, args[i])
	}
	return out
}
