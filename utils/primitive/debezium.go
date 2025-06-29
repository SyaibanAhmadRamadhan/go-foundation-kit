package primitive

// DebeziumExtractNewRecordState is a generic wrapper representing the structure of a Debezium message.
// It contains the deserialized payload and the accompanying schema.
//
// Type Parameters:
//   - T: the type of the `payload`, typically a struct representing the CDC record.
//
// Fields:
//   - Payload: the actual data payload (new record state) from Debezium
//   - Schema: the schema definition describing the structure of the payload
type DebeziumExtractNewRecordState[T any] struct {
	Payload T                                   `json:"payload"`
	Schema  DebeziumExtractNewRecordStateSchema `json:"schema"`
}

// DebeziumExtractNewRecordStatePayloadMetadata holds metadata from a Debezium payload,
// typically embedded within the record.
//
// Fields:
//   - Op: the type of operation (e.g., "c" = create, "u" = update, "d" = delete)
//   - Table: the name of the table involved
//   - LSN: the Log Sequence Number (useful for PostgreSQL replication)
//   - SourceTsMs: the timestamp (in milliseconds) of the source database change
type DebeziumExtractNewRecordStatePayloadMetadata struct {
	Op         string `json:"op"`
	Table      string `json:"table"`
	LSN        int64  `json:"lsn"`
	SourceTsMs int64  `json:"source_ts_ms"`
}

// DebeziumExtractNewRecordStateSchema represents the schema section of a Debezium message,
// describing the data types and fields in the payload.
//
// Fields:
//   - Type: the schema type (typically "struct")
//   - Optional: whether the schema itself is optional
//   - Name: the name of the schema
//   - Fields: a list of fields (columns) defined in the payload
type DebeziumExtractNewRecordStateSchema struct {
	Type     string              `json:"type"`
	Optional bool                `json:"optional"`
	Name     string              `json:"name"`
	Fields   []DebeziumFieldInfo `json:"fields"`
}

// DebeziumFieldInfo describes an individual field within a Debezium schema definition.
//
// Fields:
//   - Field: the name of the field/column
//   - Type: the Debezium/Avro type (e.g., "string", "int64")
//   - Optional: whether the field is nullable
//   - Default: the optional default value (if any)
//   - Name: an optional logical type name (e.g., "io.debezium.time.Timestamp")
//   - Version: the optional version of the field schema
type DebeziumFieldInfo struct {
	Field    string `json:"field"`
	Type     string `json:"type"`
	Optional bool   `json:"optional"`

	// Optional fields
	Default any    `json:"default,omitempty"`
	Name    string `json:"name,omitempty"`
	Version int    `json:"version,omitempty"`
}
