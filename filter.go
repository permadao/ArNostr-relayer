package relayer

import (
	"github.com/nbd-wtf/go-nostr"
)

type StorgeFilter struct {
	Cursor    string
	PageNum   int
	RelayName string
}
type ArEvent struct {
	Event  nostr.Event
	ItemId string
}
type QueryEvents struct {
	Events      []ArEvent
	Cursor      string
	HasNextPage bool
}
