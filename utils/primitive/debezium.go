package primitive

type DebeziumExtractNewRecordState[T any] struct {
	Payload T                                   `json:"payload"`
	Schema  DebeziumExtractNewRecordStateSchema `json:"schema"`
}

type DebeziumExtractNewRecordStatePayloadMetadata struct {
	Op         string `json:"op"`
	Table      string `json:"table"`
	LSN        int64  `json:"lsn"`
	SourceTsMs int64  `json:"source_ts_ms"`
}

type DebeziumExtractNewRecordStateSchema struct {
	Type     string              `json:"type"`
	Optional bool                `json:"optional"`
	Name     string              `json:"name"`
	Fields   []DebeziumFieldInfo `json:"fields"`
}

type DebeziumFieldInfo struct {
	Field    string `json:"field"`
	Type     string `json:"type"`
	Optional bool   `json:"optional"`

	// Optional fields
	Default any    `json:"default,omitempty"`
	Name    string `json:"name,omitempty"`
	Version int    `json:"version,omitempty"`
}
