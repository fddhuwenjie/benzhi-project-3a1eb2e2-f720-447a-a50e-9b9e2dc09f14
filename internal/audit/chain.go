package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func NewEvent(caseID, eventType, actorID, requestID string, revision int64, occurredAt time.Time, payload any, previousHash string) (Event, error) {
	canonical, err := CanonicalJSON(payload)
	if err != nil {
		return Event{}, err
	}
	payloadDigest := digest(canonical)
	event := Event{
		EventID: requestID + ":" + eventType, CaseID: caseID, EventType: eventType, ActorID: actorID,
		RequestID: requestID, Revision: revision, OccurredAt: occurredAt.UTC(), Payload: jsonRaw(canonical),
		PayloadDigest: payloadDigest, PreviousHash: previousHash,
	}
	event.EventHash = calculateHash(event)
	return event, nil
}

func calculateHash(event Event) string {
	input := event.EventID + "\n" + event.CaseID + "\n" + event.EventType + "\n" + event.ActorID + "\n" + event.RequestID + "\n" +
		fmt.Sprintf("%d", event.Revision) + "\n" + event.OccurredAt.UTC().Format(time.RFC3339Nano) + "\n" + event.PayloadDigest + "\n" + event.PreviousHash
	return digest([]byte(input))
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func Verify(events []Event) error {
	previous := ""
	var revision int64
	for index, event := range events {
		if index > 0 && event.Revision != revision+1 {
			return fmt.Errorf("审计链第 %d 项修订号不连续", index+1)
		}
		if event.PreviousHash != previous {
			return fmt.Errorf("审计链第 %d 项前向哈希不匹配", index+1)
		}
		canonical, err := CanonicalJSONFromRaw(event.Payload)
		if err != nil {
			return fmt.Errorf("审计链第 %d 项载荷无效: %w", index+1, err)
		}
		if digest(canonical) != event.PayloadDigest {
			return fmt.Errorf("审计链第 %d 项载荷摘要损坏", index+1)
		}
		if calculateHash(event) != event.EventHash {
			return fmt.Errorf("审计链第 %d 项事件哈希损坏", index+1)
		}
		previous, revision = event.EventHash, event.Revision
	}
	return nil
}

func CanonicalJSONFromRaw(raw []byte) ([]byte, error) {
	var value any
	if err := jsonUnmarshalNumber(raw, &value); err != nil {
		return nil, err
	}
	return CanonicalJSON(value)
}
