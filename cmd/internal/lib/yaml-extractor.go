package lib

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func ExtractFromYaml(path string) ([]string, error) {
	var matches []string
	re := regexp.MustCompile(`<path:[^>]+>`)
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !(strings.HasSuffix(p, ".yaml") || strings.HasSuffix(p, ".yml")) {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		var node yaml.Node
		if err := yaml.Unmarshal(data, &node); err != nil {
			return err
		}
		var walk func(n *yaml.Node)
		walk = func(n *yaml.Node) {
			if n.Kind == yaml.ScalarNode && n.Tag == "!!str" {
				found := re.FindAllString(n.Value, -1)
				matches = append(matches, found...)
			}
			for _, c := range n.Content {
				walk(c)
			}
		}
		walk(&node)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func SanitizePath(path string) string {
	path = strings.TrimSpace(path)
	re := regexp.MustCompile(`^<path:([^>|]+)`)
	matches := re.FindStringSubmatch(path)
	if len(matches) > 1 {
		return matches[1]
	}
	return path
}

func ProcessPath(path string) (string, string) {
	result := strings.Split(path, "#")
	if len(result) > 1 {
		return result[0], result[1]
	}
	return path, ""
}
