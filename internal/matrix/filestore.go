package matrix

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

// newFileStore creates a new filestore
func NewFileStore(dataDir string) *FileStore {
	return &FileStore{
		Path:      dataDir + "store.yaml",
		Filters:   make(map[id.UserID]string),
		NextBatch: make(map[id.UserID]string),
		Rooms:     make(map[id.RoomID]*mautrix.Room),
	}
}

// Save saves the store
func (fs *FileStore) Save() error {
	data, err := yaml.Marshal(fs)
	if err != nil {
		//fs.Logger.Error().Str("error", "Marshal err: "+err.Error())
		return err
	}
	f, err := os.Create(fs.Path)
	if err != nil {
		LOG.Error().Msg(fmt.Sprintf("Can not create file: %s", fs.Path))
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		LOG.Error().Msg("WriteFile err: " + err.Error())
	}
	return err
}

// Load loads the store
func (fs *FileStore) Load() error {
	f, err := os.Open(fs.Path)
	if err != nil {
		LOG.Error().Msg(err.Error())
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, fs)
	return err
}

// SaveFilterID sets filterID and saves
func (fs *FileStore) SaveFilterID(userID id.UserID, filterID string) {
	fs.Filters[userID] = filterID
	fs.Save()
}

// LoadFilterID loadsFilterID
func (fs *FileStore) LoadFilterID(userID id.UserID) string {
	return fs.Filters[userID]
}

// SaveNextBatch saves Next batch
func (fs *FileStore) SaveNextBatch(userID id.UserID, nextBatchToken string) {
	fs.NextBatch[userID] = nextBatchToken
	fs.Save()
}

// LoadNextBatch loads  next batch
func (fs *FileStore) LoadNextBatch(userID id.UserID) string {
	return fs.NextBatch[userID]
}

// SaveRoom saves room
func (fs *FileStore) SaveRoom(room *mautrix.Room) {
	fs.Rooms[room.ID] = room
}

// LoadRoom loads room
func (fs *FileStore) LoadRoom(roomID id.RoomID) *mautrix.Room {
	return fs.Rooms[roomID]
}
