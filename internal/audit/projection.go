package audit

func Project(events []Event) []Projection {
	items := make([]Projection, len(events))
	for i, event := range events {
		items[i] = Projection{EventID: event.EventID, EventType: event.EventType, ActorID: event.ActorID, Revision: event.Revision, OccurredAt: event.OccurredAt, EventHash: event.EventHash}
	}
	return items
}

func Head(events []Event) string {
	if len(events) == 0 {
		return ""
	}
	return events[len(events)-1].EventHash
}
