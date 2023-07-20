package main

import (
	"fmt"
	"os"
	"time"

	"github.com/eachchat/mailbot/internal/db"
	"github.com/eachchat/mailbot/internal/email"
	"github.com/eachchat/mailbot/internal/matrix"
	"github.com/eachchat/mailbot/pkg/config"
	"github.com/eachchat/mailbot/pkg/utils"
)

var tempDir = "temp/"

/*
	func viewViewHelp(roomID string, client *mautrix.Client) {
		client.SendText(id.RoomID(roomID), "Available options:\n\nmb/mailbox\t-\tViews the current used mailbox\nmbs/mailboxes\t-\tView the available mailboxes\nbl/blocklist\t-\tViews the list of blocked addresses")
	}

	func deleteTempFile(name string) {
		os.Remove(tempDir + name)
	}
*/
func main() {

	var conf config.Configuration
	var dconf db.DBCONF
	var mx matrix.MxConf
	var econf email.EmailConf
	dirPrefix := os.Getenv("BRIDGE_DATA_PATH")
	confPrefix := os.Getenv("BRIDGE_CONF_PATH")
	if len(dirPrefix) <= 0 {
		dirPrefix = "./"
	}
	if len(confPrefix) <= 0 {
		confPrefix = "./"
	}
	err := utils.MakeDir(dirPrefix)
	if err != nil {
		panic(fmt.Sprintf("%s: No such file or director.", dirPrefix))
	}
	err = utils.MakeDir(confPrefix)
	if err != nil {
		panic(fmt.Sprintf("%s: No such file or director.", confPrefix))
	}
	fmt.Println("starting to login matrix.")
	tempDir = dirPrefix + tempDir
	err = conf.InitConf(confPrefix + "config/config.yaml")
	if err != nil {
		panic(err)
	}
	logg := conf.Log
	lge := logg.New()
	dconf = conf.DB
	err = dconf.New()
	if err != nil {
		panic(fmt.Sprintf("%s can not connect.", dconf.Type))
	}

	// Init database
	err = dconf.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Can not create table, err: %s", err.Error()))
	}

	// Matrix configurations and functions
	matrix.LOG = lge
	matrix.DB = dconf.Conn
	mx.AllowedServers = conf.AllowedServers
	mx.MatrixServer = conf.MatrixServer
	mx.MatrixUserID = conf.MatrixUserID
	mx.MatrixUserPassword = conf.MatrixUserPassword
	mx.Matrixaccesstoken = conf.Matrixaccesstoken
	mx.DataDir = dirPrefix
	mx.DB = &dconf

	mx.LoginMatrix()
	mx.Client.Store = matrix.NewFileStore(mx.DataDir)
	//Email configurations and sync
	email.DB = dconf.Conn
	email.LOG = lge
	econf.MxClient = mx.Client
	email.StartMailSchedeuler(mx.Client)

	for {
		time.Sleep(1 * time.Second)
	}
}
