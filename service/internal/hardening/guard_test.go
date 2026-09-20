package hardening

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var (
	packageLevelPtrMapRe  = regexp.MustCompile(`(?m)^var\s+\w+\s+map\[[^\]]*\][*[]`)
	packageLevelMakeMapRe = regexp.MustCompile(`(?m)^var\s+\w+\s*=\s*make\(\s*map\[[^\]]*\][*[]`)
	syncMapRe             = regexp.MustCompile(`(?m)sync\.Map\b`)
)

func TestNoBusinessDataLivesInProcessMemory(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root failed: %v", err)
	}

	var violations []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "testdata" || entry.Name() == ".logs" || entry.Name() == "docs" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		clean := filepath.ToSlash(rel)
		if !strings.HasPrefix(clean, "internal/") {
			return nil
		}
		if !strings.HasSuffix(clean, ".go") || strings.HasSuffix(clean, "_test.go") {
			return nil
		}
		if strings.Contains(clean, "_mock") || strings.Contains(clean, "fixtures") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.Split(string(content), "\n")
		scan := func(re *regexp.Regexp, reason string) {
			for i, line := range lines {
				if re.MatchString(line) {
					violations = append(violations, clean+":"+strconv.Itoa(i+1)+" ("+reason+") "+strings.TrimSpace(line))
				}
			}
		}
		scan(packageLevelPtrMapRe, "package-level map holding business records")
		scan(packageLevelMakeMapRe, "package-level map holding business records")
		scan(syncMapRe, "in-process shared data store")
		return nil
	})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if len(violations) > 0 {
		t.Fatalf("in-memory business data detected (must live in the database only):\n%s", strings.Join(violations, "\n"))
	}
}
