package postgresql

import (
	"encoding/json"

	"github.com/nbd-wtf/go-nostr"
	relayer "github.com/permadao/ArNostr-relayer"
	"github.com/permadao/ArNostr-relayer/storage"
)

func (b *PostgresBackend) RestoreEvent(arEvent *relayer.ArEvent, isDelete bool) error {
	// react to different kinds of events
	evt := arEvent.Event
	itemId := arEvent.ItemId
	if evt.Kind == nostr.KindSetMetadata || evt.Kind == nostr.KindContactList || (10000 <= evt.Kind && evt.Kind < 20000) {
		// delete past events from this user
		b.DB.Exec(`DELETE FROM event WHERE pubkey = $1 AND kind = $2 AND created_at <=$3`, evt.PubKey, evt.Kind, evt.CreatedAt)
	} else if evt.Kind == nostr.KindRecommendServer {
		// delete past recommend_server events equal to this one
		b.DB.Exec(`DELETE FROM event WHERE pubkey = $1 AND kind = $2 AND content = $3 AND created_at <=$4`,
			evt.PubKey, evt.Kind, evt.Content, evt.CreatedAt)
	}

	// insert
	tagsj, _ := json.Marshal(evt.Tags)
	res, err := b.DB.Exec(`
        INSERT INTO event (id, pubkey, created_at, kind, tags, content, sig,itemid)
        VALUES ($1, $2, $3, $4, $5, $6, $7,$9)
		ON CONFLICT (id) DO UPDATE 
		SET created_at = $3, 
			kind = $4,
			tags = $5,
			content =$6,
			sig =$7,
			is_delete = $8,
			itemid = $9
		where event.created_at <$3 and  event.pubkey=$2
    `, evt.ID, evt.PubKey, evt.CreatedAt.Unix(), evt.Kind, tagsj, evt.Content, evt.Sig, isDelete, itemId)
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
