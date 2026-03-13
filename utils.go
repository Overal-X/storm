package storm

import (
	"os"
	"strings"
)

func ShellEscape(cmd string) string {
	return strings.ReplaceAll(cmd, "'", "'\\''")
}

func ChdirOrCreate(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.Chdir(dir)
}
