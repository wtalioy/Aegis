package rules

import (
	"slices"
	"sort"
	"strings"

	"aegis/pkg/utils"
)

type fileEvent struct {
	filename       string
	pathVariants   []string
	pid            uint32
	cgroupID       uint64
	matchedByInode bool
}

func (e fileEvent) hasExactPath(target string) bool {
	if target == "" {
		return false
	}
	return slices.Contains(e.pathVariants, target)
}

type pathPrefixBucket struct {
	prefix string
	rules  []*Rule
}

type fileMatcher struct {
	inodeRules map[InodeKey][]*Rule
	pathRules  map[string][]*Rule
	prefixes   []pathPrefixBucket
}

func newFileMatcher(rules []Rule) *fileMatcher {
	matcher := &fileMatcher{
		inodeRules: make(map[InodeKey][]*Rule),
		pathRules:  make(map[string][]*Rule),
		prefixes:   make([]pathPrefixBucket, 0),
	}

	prefixIndex := make(map[string]int)

	for i := range rules {
		rule := &rules[i]

		if key, ok := rule.Match.InodeKey(); ok {
			matcher.inodeRules[key] = append(matcher.inodeRules[key], rule)
		}

		if keys := rule.Match.ExactPathKeys(); len(keys) > 0 {
			for _, key := range keys {
				if key == "" {
					continue
				}
				matcher.pathRules[key] = append(matcher.pathRules[key], rule)
			}
		}

		if prefixes := rule.Match.PrefixPathKeys(); len(prefixes) > 0 {
			for _, prefix := range prefixes {
				if prefix == "" {
					continue
				}
				if idx, ok := prefixIndex[prefix]; ok {
					matcher.prefixes[idx].rules = append(matcher.prefixes[idx].rules, rule)
				} else {
					prefixIndex[prefix] = len(matcher.prefixes)
					matcher.prefixes = append(matcher.prefixes, pathPrefixBucket{
						prefix: prefix,
						rules:  []*Rule{rule},
					})
				}
			}
		}
	}

	sort.SliceStable(matcher.prefixes, func(i, j int) bool {
		return len(matcher.prefixes[i].prefix) > len(matcher.prefixes[j].prefix)
	})

	return matcher
}

func (m *fileMatcher) Match(ino, dev uint64, filename string, pid uint32, cgroupID uint64) (matched bool, rule *Rule, allowed bool) {
	if m == nil {
		return false, nil, false
	}

	variants := utils.PathVariants(filename)
	if len(variants) == 0 && filename != "" {
		if normalized := utils.NormalizeFilename(filename); normalized != "" {
			variants = append(variants, normalized)
		}
	}

	event := fileEvent{
		filename:     filename,
		pathVariants: variants,
		pid:          pid,
		cgroupID:     cgroupID,
	}

	if rules := m.inodeRules[InodeKey{Ino: ino, Dev: dev}]; len(rules) > 0 {
		inodeEvent := event
		inodeEvent.matchedByInode = true
		if matched, rule, allowed := filterRulesByAction(rules, m.matchRule, inodeEvent); matched {
			return matched, rule, allowed
		}
	}

	for _, key := range event.pathVariants {
		if key == "" {
			continue
		}
		if rules := m.pathRules[key]; len(rules) > 0 {
			if matched, rule, allowed := filterRulesByAction(rules, m.matchRule, event); matched {
				return matched, rule, allowed
			}
		}
	}

	for _, bucket := range m.prefixes {
		for _, variant := range event.pathVariants {
			if variant == "" {
				continue
			}
			if strings.HasPrefix(variant, bucket.prefix) {
				if matched, rule, allowed := filterRulesByAction(bucket.rules, m.matchRule, event); matched {
					return matched, rule, allowed
				}
				break
			}
		}
	}

	return false, nil, false
}

func (m *fileMatcher) matchRule(rule *Rule, event fileEvent) bool {
	match := rule.Match
	if match.Filename == "" && len(match.PrefixPathKeys()) == 0 {
		return false
	}

	// 1) Exact path keys (skip if matched by inode already)
	if len(match.ExactPathKeys()) > 0 && !event.matchedByInode {
		found := slices.ContainsFunc(match.ExactPathKeys(), event.hasExactPath)
		// If keys include directory components (slash), require exact-variant match.
		// If keys are all basenames, allow fallback to basename matching below.
		hasSlashKey := slices.ContainsFunc(match.ExactPathKeys(), func(s string) bool { return strings.Contains(s, "/") })
		if !found && hasSlashKey {
			return false
		}
	}

	// 2) Prefix directory keys
	if len(match.PrefixPathKeys()) > 0 {
		found := false
		for _, prefix := range match.PrefixPathKeys() {
			if prefix == "" {
				continue
			}
			for _, variant := range event.pathVariants {
				if variant == "" {
					continue
				}
				if strings.HasPrefix(ensureTrailingSlash(variant), prefix) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// 3) If exact keys are bare filenames (no slash), allow basename matching.
	if len(match.ExactPathKeys()) > 0 {
		allBasenames := true
		for _, k := range match.ExactPathKeys() {
			if strings.Contains(k, "/") {
				allBasenames = false
				break
			}
		}
		if allBasenames {
			baseMatch := false
			for _, variant := range event.pathVariants {
				if variant == "" {
					continue
				}
				vk := pathBase(variant)
				for _, key := range match.ExactPathKeys() {
					if vk == key {
						baseMatch = true
						break
					}
				}
				if baseMatch {
					break
				}
			}
			// Also check raw filename
			if !baseMatch {
				vk := pathBase(event.filename)
				for _, key := range match.ExactPathKeys() {
					if vk == key {
						baseMatch = true
						break
					}
				}
			}
			if !baseMatch {
				return false
			}
		}
	}

	return matchCgroupID(match.CgroupID, event.cgroupID) && matchPID(match.PID, event.pid)
}

// pathBase is a minimal, allocation-free base path extractor for both absolute and relative paths.
func pathBase(p string) string {
	if p == "" {
		return ""
	}
	// Find last slash
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			if i == len(p)-1 {
				return ""
			}
			return p[i+1:]
		}
	}
	return p
}

func ensureTrailingSlash(path string) string {
	if path == "" || path == "/" || strings.HasSuffix(path, "/") {
		return path
	}
	return path + "/"
}
