package databases

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONB[T any] []byte

func (j *JSONB[T]) Scan(value any) error {
	if j == nil {
		return fmt.Errorf("JSONB.Scan: receiver is nil")
	}
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		buf := make([]byte, len(v))
		copy(buf, v)
		*j = JSONB[T](buf)
		return nil
	case string:
		buf := []byte(v)
		*j = JSONB[T](buf)
		return nil
	default:
		return fmt.Errorf("JSONB.Scan: unsupported source type %T", value)
	}
}

func (j JSONB[T]) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *JSONB[T]) MarshalFrom(v T) error {
	if j == nil {
		return fmt.Errorf("JSONB.MarshalFrom: receiver is nil")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	*j = JSONB[T](b)
	return nil
}

func (j JSONB[T]) Unmarshal(dst *T) error {
	if len(j) == 0 {
		var zero T
		*dst = zero
		return nil
	}
	return json.Unmarshal(j, dst)
}

func (j JSONB[T]) Into() (T, error) {
	var out T
	if len(j) == 0 {
		return out, nil
	}
	err := json.Unmarshal(j, &out)
	return out, err
}

func (j JSONB[T]) IsZero() bool {
	return len(j) == 0
}
