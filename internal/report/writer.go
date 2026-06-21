package report

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Write(content string) (string, error) {
	dir := "reports"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s.md", time.Now().Format("2006-01-02"))
	path := filepath.Join(dir, filename)

	header := fmt.Sprintf("# Daily Briefing - %s\n\n", time.Now().Format("Monday, January 2, 2006"))

	if err := os.WriteFile(path, []byte(header+content), 0644); err != nil {
		return "", err
	}

	return path, nil
}
