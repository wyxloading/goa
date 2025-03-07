package helper

import (
	"bufio"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

const goModFileName = "go.mod"

// GetModuleName returns the module name from a go.mod
func GetModuleName(rootDir string) (string, error) {
	goModPath := filepath.Join(rootDir, goModFileName)
	if _, err := os.Stat(goModPath); err != nil {
		return "", err
	}

	f, err := os.Open(goModPath)

	if err != nil {
		return "", err
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		trimmedLine := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(trimmedLine, "module") {
			return trimmedLine[7:], nil
		}
	}

	return "", errors.New("module name cannot be found")
}
