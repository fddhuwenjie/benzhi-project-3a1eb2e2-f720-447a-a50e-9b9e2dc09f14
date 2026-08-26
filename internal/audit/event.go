package audit

import "time"

type Event struct {
	EventID       string    `json:"event_id"`
	CaseID        string    `json:"case_id"`
	EventType     string    `json:"event_type"`
	ActorID       string    `json:"actor_id"`
	RequestID     string    `json:"request_id"`
	Revision      int64     `json:"revision"`
	OccurredAt    time.Time `json:"occurred_at"`
	Payload       jsonRaw   `json:"payload"`
	PayloadDigest string    `json:"payload_digest"`
	PreviousHash  string    `json:"previous_hash"`
	EventHash     string    `json:"event_hash"`
}

type jsonRaw []byte

func (r jsonRaw) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}
	return r, nil
}

func (r *jsonRaw) UnmarshalJSON(data []byte) error {
	*r = append((*r)[:0], data...)
	return nil
}

func (e Event) PayloadJSON() []byte { return append([]byte(nil), e.Payload...) }

type Projection struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	ActorID    string    `json:"actor_id"`
	Revision   int64     `json:"revision"`
	OccurredAt time.Time `json:"occurred_at"`
	EventHash  string    `json:"event_hash"`
}
