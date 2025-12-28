package prompt

const AnalyzeSystemPrompt = `You are Aegis AI's security analyst. Analyze entities and provide actionable insights.

Structure:
1. **Current State**: Brief overview
2. **Anomalies**: Unusual patterns (file access, network, process behavior)
3. **Baseline**: Compare metrics (note if baseline unavailable)
4. **Security Assessment**: Likelihood (low/medium/high), risk level, reasoning
5. **Recommendations**: Immediate actions, investigation steps, rule suggestions

Use markdown. Be concise but thorough.`

const AnalyzeUserTemplate = `Analysis request:
- **Type**: {{.Type}}
- **ID**: {{.ID}}

{{if eq .Type "process"}}
**Process Details**:
- PID: {{.PID}}
- Command: {{.CommandLine}}
- Start time: {{.StartTime}}
- File operations: {{.FileOpenCount}} total
- Network connections: {{.NetConnectCount}} total

Analyze this process for security concerns and anomalous behavior.
{{end}}

{{if eq .Type "workload"}}
**Workload Details**:
- Cgroup path: {{.CgroupPath}}
- Active processes: {{.ProcessCount}}
- Total events: {{.TotalEvents}}

Analyze this workload for security posture and activity patterns.
{{end}}

Provide a comprehensive security analysis.`
