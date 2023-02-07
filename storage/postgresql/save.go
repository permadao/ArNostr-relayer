package postgresql

import (
	"encoding/json"

	"github.com/nbd-wtf/go-nostr"
	"github.com/permadao/ArNostr-relayer/storage"
)

func (b *PostgresBackend) SaveEvent(evt *nostr.Event) error {
	// react to different kinds of events
	if evt.Kind == nostr.KindSetMetadata || evt.Kind == nostr.KindContactList || (10000 <= evt.Kind && evt.Kind < 20000) {
		// delete past events from this user
		b.DB.Exec(`DELETE FROM event WHERE pubkey = $1 AND kind = $2`, evt.PubKey, evt.Kind)
	} else if evt.Kind == nostr.KindRecommendServer {
		// delete past recommend_server events equal to this one
		b.DB.Exec(`DELETE FROM event WHERE pubkey = $1 AND kind = $2 AND content = $3`,
			evt.PubKey, evt.Kind, evt.Content)
	}

	// insert
	tagsj, _ := json.Marshal(evt.Tags)
	res, err := b.DB.Exec(`
        INSERT INTO event (id, pubkey, created_at, kind, tags, content, sig)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
    `, evt.ID, evt.PubKey, evt.CreatedAt.Unix(), evt.Kind, tagsj, evt.Content, evt.Sig)
	if err != nil {
		return err
	}

	nr, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if nr == 0 {
		return storage.ErrDupEvent
	}

	return nil
}
