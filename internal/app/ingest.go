package app

import (
	"aegis/internal/platform/events"
	"aegis/internal/policy"
	"aegis/internal/shared/stream"
	"aegis/internal/system"
	"aegis/internal/telemetry"
)

type IngestPipeline struct {
	telemetry   *telemetry.Service
	policy      *policy.Service
	stats       *system.Stats
	eventStream *stream.Hub[telemetry.Event]
	alertStream *stream.Hub[system.Alert]
}

func NewIngestPipeline(
	telemetryService *telemetry.Service,
	policyService *policy.Service,
	stats *system.Stats,
	eventStream *stream.Hub[telemetry.Event],
	alertStream *stream.Hub[system.Alert],
) *IngestPipeline {
	return &IngestPipeline{
		telemetry:   telemetryService,
		policy:      policyService,
		stats:       stats,
		eventStream: eventStream,
		alertStream: alertStream,
	}
}

func (p *IngestPipeline) ProcessRawSample(data []byte) (*telemetry.Event, policy.Decision, error) {
	decoded, err := events.DecodeSample(data)
	if err != nil {
		return nil, policy.Decision{}, err
	}
	record, err := p.telemetry.Ingest(decoded)
	if err != nil {
		return nil, policy.Decision{}, err
	}
	decision := p.ProcessRecord(record)
	return &record.Event, decision, nil
}

func (p *IngestPipeline) ProcessRecord(record *telemetry.Record) policy.Decision {
	if record == nil {
		return policy.Decision{Type: policy.DecisionNoMatch}
	}
	event := &record.Event

	switch event.Type {
	case telemetry.EventTypeExec:
		p.stats.RecordExec()
	case telemetry.EventTypeFile:
		p.stats.RecordFile()
	case telemetry.EventTypeConnect:
		p.stats.RecordConnect()
	}

	p.eventStream.Publish(*event)

	decision := p.policy.Evaluate(record)
	for _, alert := range decision.Alerts {
		p.stats.AddAlert(alert)
		p.telemetry.RecordAlert(event.CgroupID, alert.Blocked)
		p.alertStream.Publish(alert)
	}

	return decision
}
