package rules

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestInodeMatchingWithHardlink(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "sensitive.txt")
	alias := filepath.Join(dir, "alias.txt")

	if err := os.WriteFile(original, []byte("top secret"), 0o600); err != nil {
		t.Fatalf("Failed to create original file: %v", err)
	}
	if err := os.Link(original, alias); err != nil {
		t.Fatalf("Failed to create hardlink: %v", err)
	}

	rules := []Rule{
		{
			Name:     "Alert on sensitive file",
			Severity: "high",
			Action:   ActionAlert,
			Match: MatchCondition{
				Filename: original,
			},
		},
	}

	engine := NewEngine(rules)

	info, err := os.Stat(alias)
	if err != nil {
		t.Fatalf("Failed to stat hardlink: %v", err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatal("expected Stat_t for hardlink")
	}

	matched, rule, allowed := engine.MatchFile(stat.Ino, uint64(stat.Dev), alias, 0, 0)
	if !matched {
		t.Fatal("Expected match for inode")
	}
	if allowed {
		t.Fatal("Expected alert rule to block/alert, not allow")
	}
	if rule == nil || rule.Name != "Alert on sensitive file" {
		t.Fatalf("Unexpected rule returned: %+v", rule)
	}
}

func TestFileRuleFallsBackToPathWhenInodeMissing(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "nonexistent.txt")

	rules := []Rule{
		{
			Name:     "Monitor missing file",
			Severity: "medium",
			Action:   ActionAlert,
			Match: MatchCondition{
				Filename: target,
			},
		},
	}

	engine := NewEngine(rules)

	matched, rule, allowed := engine.MatchFile(0, 0, target, 0, 0)
	if !matched {
		t.Fatal("Expected path-based match even when inode missing")
	}
	if allowed {
		t.Fatal("Expected alert action to block/alert")
	}
	if rule == nil || rule.Name != "Monitor missing file" {
		t.Fatalf("Unexpected rule returned: %+v", rule)
	}
}

func TestRelativePathRuleMatches(t *testing.T) {
	rules := []Rule{
		{
			Name:     "Docs file alert",
			Severity: "low",
			Action:   ActionAlert,
			Match: MatchCondition{
				Filename: "docs/readme.md",
			},
		},
	}

	engine := NewEngine(rules)

	matched, _, _ := engine.MatchFile(0, 0, "docs/readme.md", 0, 0)
	if !matched {
		t.Fatal("Expected relative filename rule to match")
	}
}

func TestWildcardFilenameMatchesCanonicalForms(t *testing.T) {
	rules := []Rule{
		{
			Name:     "Monitor log dir",
			Severity: "medium",
			Action:   ActionAlert,
			Match: MatchCondition{
				Filename: "/var/log/*",
			},
		},
	}

	engine := NewEngine(rules)

	if matched, _, _ := engine.MatchFile(0, 0, "var/log/app.log", 0, 0); !matched {
		t.Fatal("Expected wildcard rule to match relative form")
	}

	if matched, _, _ := engine.MatchFile(0, 0, "/var/log/app.log", 0, 0); !matched {
		t.Fatal("Expected wildcard rule to match canonical form")
	}
}
