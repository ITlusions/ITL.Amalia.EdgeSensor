package event

import (
	"encoding/json"
	"time"
)

type Event struct {
	SchemaVersion string          `json:"schema_version"`
	SourceType    string          `json:"source_type"`
	DeviceID      string          `json:"device_id"`
	ObservedAt    time.Time       `json:"observed_at"`
	ReceivedAt    time.Time       `json:"received_at"`
	Severity      string          `json:"severity"`
	Data          json.RawMessage `json:"data"`
}

func New(sourceType, deviceID, severity string, observedAt time.Time, data any) (Event, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return Event{}, err
	}

	return Event{
		SchemaVersion: "1.0",
		SourceType:    sourceType,
		DeviceID:      deviceID,
		ObservedAt:    observedAt.UTC(),
		ReceivedAt:    time.Now().UTC(),
		Severity:      severity,
		Data:          payload,
	}, nil
}
