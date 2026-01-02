package rules

import (
	"aegis/pkg/events"
	"aegis/pkg/utils"
	"time"
)

type connectMatcher struct {
	rules         []*Rule
	testingBuffer *TestingBuffer
}

func newConnectMatcher(rules []Rule, testingBuffer *TestingBuffer) *connectMatcher {
	matcher := &connectMatcher{
		rules:         make([]*Rule, 0),
		testingBuffer: testingBuffer,
	}
	for i := range rules {
		if rules[i].Match.DestPort != 0 || rules[i].Match.DestIP != "" {
			matcher.rules = append(matcher.rules, &rules[i])
		}
	}
	return matcher
}

func (m *connectMatcher) Match(event *events.ConnectEvent) (matched bool, rule *Rule, allowed bool) {
	return filterRulesByAction(m.rules, m.matchRule, event)
}

func (m *connectMatcher) CollectAlerts(event *events.ConnectEvent, processName string) []MatchedAlert {
	var alerts []MatchedAlert
	seen := make(map[*Rule]bool)

	for _, rule := range m.rules {
		if seen[rule] {
			continue
		}
		if m.matchRule(rule, event) {
			seen[rule] = true
			if rule.IsTesting() {
				if m.testingBuffer != nil {
					hit := &TestingHit{
						RuleName:    rule.Name,
						HitTime:     time.Now(),
						EventType:   events.EventTypeConnect,
						EventData:   event,
						PID:         event.Hdr.PID,
						ProcessName: processName,
					}
					m.testingBuffer.RecordHit(hit)
				}
			} else {
				alerts = append(alerts, MatchedAlert{
					Rule:    *rule,
					Message: rule.Description,
				})
			}
		}
	}
	return alerts
}

func (m *connectMatcher) matchRule(rule *Rule, event *events.ConnectEvent) bool {
	match := rule.Match
	if match.DestPort == 0 && match.DestIP == "" {
		return false
	}
	if match.DestPort != 0 && event.Port != match.DestPort {
		return false
	}
	if match.DestIP != "" {
		if eventIP := utils.ExtractIP(event); eventIP == "" || !match.MatchIP(eventIP) {
			return false
		}
	}
	return matchCgroupID(match.CgroupID, event.Hdr.CgroupID) && matchPID(match.PID, event.Hdr.PID)
}
