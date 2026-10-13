package domain

import "time"

// Event types published through EventPublisher.
const (
	// EventShelfDomainChanged: a shelf's domain was set, changed or cleared, or the shelf or its owner was deleted.
	EventShelfDomainChanged = "shelf.domain.changed"
)

// EventPublisher sends best-effort notifications to external services. A receiver
// must never depend on them, the API is the truth.
type EventPublisher interface {
	Publish(eventType string, data any)
}

// shelfDomainChangedData is the payload of EventShelfDomainChanged: ids and domains only.
type shelfDomainChangedData struct {
	ShelfId        string `json:"shelf_id"`
	OwnerId        string `json:"owner_id"`
	Domain         string `json:"domain"`
	PreviousDomain string `json:"previous_domain"`
}

func (s *Service) publish(eventType string, data any) {
	if s != nil && s.Events != nil {
		s.Events.Publish(eventType, data)
	}
}

func (s *Service) publishDomainChanged(shelfId, ownerId, domain, previous string) {
	if domain == previous {
		return
	}
	s.publish(EventShelfDomainChanged, shelfDomainChangedData{ShelfId: shelfId, OwnerId: ownerId, Domain: domain, PreviousDomain: previous})
}

func unixToRfc3339(seconds int64) string {
	return time.Unix(seconds, 0).UTC().Format(time.RFC3339)
}
