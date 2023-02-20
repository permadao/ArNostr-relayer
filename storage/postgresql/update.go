package postgresql

import (
	"encoding/json"

	"github.com/nbd-wtf/go-nostr"
	"github.com/permadao/ArNostr-relayer/storage"
)

func (b *PostgresBackend) UpdateItemId(evt *nostr.Event, itemid string) error {
	// insert or update itemid
	tagsj, _ := json.Marshal(evt.Tags)
	res, err := b.DB.Exec(`
        INSERT INTO event (id, pubkey, created_at, kind, tags, content, sig, itemid)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE 
		SET itemid = $8
    `, evt.ID, evt.PubKey, evt.CreatedAt.Unix(), evt.Kind, tagsj, evt.Content, evt.Sig, itemid)
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
