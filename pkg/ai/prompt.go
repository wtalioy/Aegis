package ai

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"eulerguard/pkg/types"
)

const DiagnosisSystemPrompt = `You are EulerGuard AI, an expert Linux kernel security analyst. 
You analyze eBPF telemetry data to diagnose system issues and security threats.
Be concise, technical, and actionable. Use markdown formatting.
Focus on: root cause, security implications, and remediation steps.`

const ChatSystemPrompt = `You are EulerGuard AI, an intelligent assistant for Linux kernel security monitoring.

Your capabilities:
- Analyze real-time eBPF telemetry (process execution, file access, network connections)
- Explain security alerts and detection rules
- Provide remediation guidance for security issues
- Answer questions about system behavior and kernel security

Guidelines:
- Be conversational but technically accurate
- Use markdown for formatting (headers, code blocks, lists)
- Reference specific data from the context when relevant
- If asked about something not in the context, say so
- Keep responses focused and under 300 words unless more detail is requested

You have access to live system telemetry that updates with each message.`

const DiagnosisTemplateText = `## System Diagnosis Request

### Current System State
- **Load Level**: {{.LoadLevel}} (Exec: {{.ExecRate}}/s, File: {{.FileRate}}/s, Network: {{.NetworkRate}}/s)
- **Tracked Processes**: {{.ProcessCount}}
- **Active Workloads**: {{.WorkloadCount}}
- **Total Alerts**: {{.AlertCount}}

{{if .TopWorkloads}}
### Most Active Workloads
{{range .TopWorkloads}}
- {{.CgroupPath}} ({{.TotalEvents}} events, {{.AlertCount}} alerts)
{{end}}
{{end}}

{{if .RecentAlerts}}
### Recent Security Alerts
{{range .RecentAlerts}}
- **{{.RuleName}}** ({{.Severity}}): Process "{{.ProcessName}}" {{if .WasBlocked}}[BLOCKED]{{end}} - {{.Count}} occurrence(s)
{{end}}
{{end}}

{{if .RecentProcesses}}
### Recent Process Tree Activity (parent → child)
{{range .RecentProcesses}}
- {{.ParentComm}} → {{.Comm}}{{if .Blocked}} [BLOCKED]{{end}} (×{{.Count}})
{{end}}
{{end}}

{{if .RecentConnections}}
### Recent Network Connections
{{range .RecentConnections}}
- {{.Destination}}{{if .Blocked}} [BLOCKED]{{end}} (×{{.Count}})
{{end}}
{{end}}

{{if .RecentFileAccess}}
### Recent File Access
{{range .RecentFileAccess}}
- {{.Path}}{{if .Blocked}} [BLOCKED]{{end}} (×{{.Count}})
{{end}}
{{end}}

---

Please analyze this telemetry data and provide:
1. **Diagnosis**: What is happening on this system?
2. **Security Assessment**: Are there any security concerns?
3. **Recommendations**: What actions should be taken?

Keep your response concise and actionable (under 400 words).`

const ChatContextTemplate = `[LIVE SYSTEM CONTEXT - Updated at {{.Timestamp.Format "15:04:05"}}]

System Load: {{.LoadLevel | upper}}
├─ Process Executions: {{.ExecRate}}/s
├─ File Operations: {{.FileRate}}/s  
└─ Network Connections: {{.NetworkRate}}/s

Summary: {{.ProcessCount}} processes tracked, {{.WorkloadCount}} workloads, {{.AlertCount}} total alerts
{{if .TopWorkloads}}
Active Workloads:
{{range .TopWorkloads}}- {{.CgroupPath}} ({{.TotalEvents}} events{{if gt .AlertCount 0}}, {{.AlertCount}} alerts{{end}})
{{end}}{{end}}{{if .RecentAlerts}}
Security Alerts:
{{range .RecentAlerts}}- [{{.Severity | upper}}] {{.RuleName}}: "{{.ProcessName}}" {{if .WasBlocked}}BLOCKED{{else}}logged{{end}} (×{{.Count}})
{{end}}{{end}}{{if .RecentProcesses}}
Process Activity (parent→child):
{{range .RecentProcesses}}- {{.ParentComm}}→{{.Comm}}{{if .Blocked}} BLOCKED{{end}} (×{{.Count}})
{{end}}{{end}}{{if .RecentConnections}}
Network Destinations:
{{range .RecentConnections}}- {{.Destination}}{{if .Blocked}} BLOCKED{{end}} (×{{.Count}})
{{end}}{{end}}{{if .RecentFileAccess}}
File Access:
{{range .RecentFileAccess}}- {{.Path}}{{if .Blocked}} BLOCKED{{end}} (×{{.Count}})
{{end}}{{end}}
[END CONTEXT]`

var (
	diagnosisTmpl   *template.Template
	chatContextTmpl *template.Template
)

func init() {
	funcs := template.FuncMap{
		"upper": strings.ToUpper,
	}
	diagnosisTmpl = template.Must(template.New("diagnosis").Parse(DiagnosisTemplateText))
	chatContextTmpl = template.Must(template.New("chatContext").Funcs(funcs).Parse(ChatContextTemplate))
}

type PromptContext struct {
	types.SystemSnapshot
}

func GeneratePrompt(snapshot types.SystemSnapshot) (string, error) {
	ctx := PromptContext{
		SystemSnapshot: snapshot,
	}
	var buf bytes.Buffer
	if err := diagnosisTmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func FormatSnapshotSummary(snapshot types.SystemSnapshot) string {
	var parts []string
	parts = append(parts, "Load: "+snapshot.LoadLevel)
	if len(snapshot.RecentAlerts) > 0 {
		parts = append(parts, "Alerts: "+formatAlertsSummary(snapshot.RecentAlerts))
	}
	if len(snapshot.TopWorkloads) > 0 {
		parts = append(parts, "Top: "+snapshot.TopWorkloads[0].CgroupPath)
	}
	return strings.Join(parts, " | ")
}

func formatAlertsSummary(alerts []types.AlertSummary) string {
	if len(alerts) == 0 {
		return "none"
	}

	var critCount, highCount int
	for _, a := range alerts {
		switch a.Severity {
		case "critical":
			critCount += a.Count
		case "high":
			highCount += a.Count
		}
	}

	if critCount > 0 {
		return "critical"
	}
	if highCount > 0 {
		return "high"
	}
	return alerts[0].RuleName
}

func FormatContextForChat(snapshot types.SystemSnapshot) string {
	var buf bytes.Buffer
	err := chatContextTmpl.Execute(&buf, snapshot)
	if err != nil {
		return fmt.Sprintf("[System: %s load, %d alerts]",
			snapshot.LoadLevel, snapshot.AlertCount)
	}
	return buf.String()
}
