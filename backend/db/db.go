package db

import (
	"errors"
	"sort"
	"sync"

	"example.com/csiproject/backend/api"
)

var (
	ErrDoesNotExist  = errors.New("does not exist")
	ErrAlreadyExists = errors.New("already exists")
)

type DeleteResponse struct {
	ID string `json:"ID"`
}

// Database is the interface used by the server to load and store volumes.
type Database interface {
	// GetVolumes returns a copy of all volumes, sorted by ID.
	GetVolumes() ([]api.Volume, error)

	// GetVolumesByID returns a single volume by ID, or ErrDoesNotExist if
	// an volume with that ID does not exist.
	GetVolumeByID(id string) (api.Volume, error)

	// DeleteVolumesByID returns ErrDoesNotExist if
	// an volume with that ID does not exist, otherwise deletes the volume entry.
	DeleteVolumeByID(id string) (DeleteResponse, error)

	// AddVolume adds a single volume, or ErrAlreadyExists if an volume with
	// the given ID already exists.
	AddVolume(volume api.Volume) error
}

// MemoryDatabase is a Database implementation that uses a simple
// in-memory map to store the volumes.
type MemoryDatabase struct {
	lock    sync.RWMutex
	volumes map[string]api.Volume
}

// NewMemoryDatabase creates a new in-memory database.
func NewMemoryDatabase() *MemoryDatabase {
	return &MemoryDatabase{volumes: make(map[string]api.Volume)}
}

func (d *MemoryDatabase) GetVolumes() ([]api.Volume, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	// Make a copy of the volumes map (as a slice)
	volumes := make([]api.Volume, 0, len(d.volumes))
	for _, volume := range d.volumes {
		volumes = append(volumes, volume)
	}

	// Sort by ID so we return them in a defined order
	sort.Slice(volumes, func(i, j int) bool {
		return volumes[i].ID < volumes[j].ID
	})
	return volumes, nil
}

func (d *MemoryDatabase) GetVolumeByID(id string) (api.Volume, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	volume, ok := d.volumes[id]
	if !ok {
		return api.Volume{}, ErrDoesNotExist
	}
	return volume, nil
}

func (d *MemoryDatabase) AddVolume(volume api.Volume) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if _, ok := d.volumes[volume.ID]; ok {
		return ErrAlreadyExists
	}
	d.volumes[volume.ID] = volume
	return nil
}

func (d *MemoryDatabase) DeleteVolumeByID(id string) (DeleteResponse, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	_, ok := d.volumes[id]
	if !ok {
		return DeleteResponse{}, ErrDoesNotExist
	}
	delete(d.volumes, id)
	return DeleteResponse{ID: id}, nil
}
