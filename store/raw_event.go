package store

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"

	"github.com/nbd-wtf/go-nostr"
	nostr_binary "github.com/nbd-wtf/go-nostr/binary"
	"golang.org/x/exp/mmap"
)

type RawEventStore struct {
	Path string

	lastOffset   int64
	writeHandler *os.File
	readHandler  *mmap.ReaderAt
	writeMutex   sync.Mutex
	pendingRemap bool
}

func (r *RawEventStore) Init() error {
	var err error

	r.writeHandler, err = os.OpenFile(r.Path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("can't open file for writing: %w", err)
	}

	if err := r.remap(); err != nil {
		return err
	}

	r.lastOffset = int64(r.readHandler.Len())

	return nil
}

func (r *RawEventStore) Deinit() {
	r.writeHandler.Close()
	r.readHandler.Close()
}

func (r *RawEventStore) SaveEvent(evt *nostr.Event) (offset int64, err error) {
	r.writeMutex.Lock()
	defer r.writeMutex.Unlock()

	data, _ := nostr_binary.Marshal(evt)
	offset = r.lastOffset

	s := uint16(len(data))
	var sizeb [2]byte
	binary.LittleEndian.PutUint16(sizeb[:], s)
	r.writeHandler.Write(sizeb[:])

	n, err := r.writeHandler.Write(data)
	if err != nil {
		return 0, fmt.Errorf("failed to write: %w", err)
	}
	r.lastOffset += 2 + int64(n)
	r.pendingRemap = true

	return offset, nil
}

func (r *RawEventStore) ReadEvent(offset int64, evt *nostr.Event) error {
	if r.pendingRemap {
		r.remap()
	}

	var sizeb [2]byte
	if _, err := r.readHandler.ReadAt(sizeb[:], offset); err != nil {
		return fmt.Errorf("failed to read size at %d: %w", offset, err)
	}
	s := binary.LittleEndian.Uint16(sizeb[:])
	bin := make([]byte, s)
	if _, err := r.readHandler.ReadAt(bin, offset+2); err != nil {
		return fmt.Errorf("failed to read event at %d: %w", offset, err)
	}

	return nostr_binary.Unmarshal(bin, evt)
}

func (r *RawEventStore) remap() error {
	var err error
	r.readHandler, err = mmap.Open("events")
	if err != nil {
		return fmt.Errorf("can't mmap file: %w", err)
	}
	return nil
}
