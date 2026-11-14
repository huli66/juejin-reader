package crawl

import (
	"encoding/json"
	"os"
	"regexp"
)

var unsafeChars = regexp.MustCompile(`[\\/:*?<>|\s]+`)

func safeName(s string) string {
	return unsafeChars.ReplaceAllString(s, "_")
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func saveJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
