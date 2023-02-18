package relayer

import (
	"github.com/nbd-wtf/go-nostr"
)

type StorgeFilter struct {
	Cursor    string
	PageNum   int8
	RelayName string
}
type QueryEvents struct {
	Events      []nostr.Event
	Cursor      string
	HasNextPage bool
}
