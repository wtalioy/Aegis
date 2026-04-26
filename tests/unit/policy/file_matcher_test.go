package policy_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"aegis/internal/policy"
	"aegis/internal/policy/rules"
	"aegis/tests/helpers"
)

func TestFileMatcher_MatchesHardlinksByInode(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "sensitive.txt")
	alias := filepath.Join(dir, "alias.txt")

	if err := os.WriteFile(original, []byte("top secret"), 0o600); err != nil {
		t.Fatalf("failed to create original file: %v", err)
	}
	if err := os.Link(original, alias); err != nil {
		t.Fatalf("failed to create hardlink: %v", err)
	}

	engine := rules.NewEngine([]policy.Rule{
		helpers.ActiveFileRule("Alert on sensitive file", original, policy.ActionAlert),
	})

	info, err := os.Stat(alias)
	if err != nil {
		t.Fatalf("failed to stat hardlink: %v", err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatal("expected Stat_t for hardlink")
	}

	matched, rule, allowed := engine.MatchFile(stat.Ino, uint64(stat.Dev), alias, 0, 0)
	if !matched {
		t.Fatal("expected match for inode")
	}
	if allowed {
		t.Fatal("expected alert rule to remain non-allowing")
	}
	if rule == nil || rule.Name != "Alert on sensitive file" {
		t.Fatalf("unexpected rule returned: %+v", rule)
	}
}

func TestFileMatcher_FallsBackToPathWhenInodeIsMissing(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "nonexistent.txt")

	engine := rules.NewEngine([]policy.Rule{
		helpers.ActiveFileRule("Monitor missing file", target, policy.ActionAlert),
	})

	matched, rule, allowed := engine.MatchFile(0, 0, target, 0, 0)
	if !matched {
		t.Fatal("expected path-based match even when inode missing")
	}
	if allowed {
		t.Fatal("expected alert action to remain non-allowing")
	}
	if rule == nil || rule.Name != "Monitor missing file" {
		t.Fatalf("unexpected rule returned: %+v", rule)
	}
}

func TestFileMatcher_MatchesRelativePaths(t *testing.T) {
	engine := rules.NewEngine([]policy.Rule{
		helpers.ActiveFileRule("Docs file alert", "docs/readme.md", policy.ActionAlert),
	})

	matched, _, _ := engine.MatchFile(0, 0, "docs/readme.md", 0, 0)
	if !matched {
		t.Fatal("expected relative filename rule to match")
	}
}

func TestFileMatcher_MatchesWildcardRulesAgainstRelativeAndCanonicalForms(t *testing.T) {
	engine := rules.NewEngine([]policy.Rule{
		helpers.ActiveFileRule("Monitor log dir", "/var/log/*", policy.ActionAlert),
	})

	if matched, _, _ := engine.MatchFile(0, 0, "var/log/app.log", 0, 0); !matched {
		t.Fatal("expected wildcard rule to match relative form")
	}
	if matched, _, _ := engine.MatchFile(0, 0, "/var/log/app.log", 0, 0); !matched {
		t.Fatal("expected wildcard rule to match canonical form")
	}
}
