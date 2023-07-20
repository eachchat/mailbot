package matrix

import (
	"github.com/eachchat/mailbot/internal/db"
	"github.com/eachchat/mailbot/pkg/config"
	"gorm.io/gorm"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

var DB *gorm.DB

// MxConf the confingurations of matrix
type MxConf struct {
	config.MatrixConf
	FileStore
	Client  *mautrix.Client
	DataDir string
	DB      *db.DBCONF
}

// FileStore required by the bridgeAPI
type FileStore struct {
	Path      string
	Filters   map[id.UserID]string        `yaml:"filter_id"`
	NextBatch map[id.UserID]string        `yaml:"next_batch"`
	Rooms     map[id.RoomID]*mautrix.Room `yaml:"-"`
}
