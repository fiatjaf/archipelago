package main

import (
	"fmt"

	"github.com/fiatjaf/archipelago/store"
	"github.com/nbd-wtf/go-nostr"
)

func main() {
	res := store.RawEventStore{Path: "events"}
	res.Init()
	defer res.Deinit()

	var offset1 int64
	var offset2 int64

	{
		evt := &nostr.Event{
			Content:   "bla",
			Tags:      nostr.Tags{},
			CreatedAt: nostr.Now(),
			Kind:      1,
		}
		sk := nostr.GeneratePrivateKey()
		evt.Sign(sk)
		offset, err := res.SaveEvent(evt)
		fmt.Println("written", offset, err)
		offset1 = offset
	}
	{
		evt := &nostr.Event{
			Content:   "bli",
			Tags:      nostr.Tags{nostr.Tag{"t", "spam"}},
			CreatedAt: nostr.Now(),
			Kind:      1,
		}
		sk := nostr.GeneratePrivateKey()
		evt.Sign(sk)
		offset, err := res.SaveEvent(evt)
		fmt.Println("written", offset, err)
		offset2 = offset
	}

	{
		var evtr nostr.Event
		res.ReadEvent(offset2, &evtr)
		fmt.Println(evtr)
	}
	{
		var evtr nostr.Event
		res.ReadEvent(offset1, &evtr)
		fmt.Println(evtr)
	}
	{
		var evtr nostr.Event
		res.ReadEvent(0, &evtr)
		fmt.Println(evtr)
	}
	{
		var evtr nostr.Event
		res.ReadEvent(284, &evtr)
		fmt.Println(evtr)
	}
}
