package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/eachchat/mailbot/internal/db"
	"github.com/eachchat/mailbot/internal/email"
	"github.com/eachchat/mailbot/internal/matrix"
	"github.com/eachchat/mailbot/pkg/config"
	"github.com/eachchat/mailbot/pkg/utils"
)

var tempDir = "temp/"

func main() {
	done := make(chan struct{})
	q := make(chan os.Signal, 1)
	signal.Notify(q, os.Interrupt, syscall.SIGABRT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
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
	dirPrefix = utils.CheckeDir(dirPrefix)

	confPrefix = utils.CheckeDir(confPrefix)

	fmt.Println("starting to login matrix.")
	tempDir = dirPrefix + tempDir
	err := conf.InitConf(confPrefix + "config.yaml")
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
	mx.DefaultMailCheckInterval = conf.DefaultMailCheckInterval
	mx.DataDir = dirPrefix
	mx.DB = &dconf

	mx.LoginMatrix()
	mx.Client.Store = matrix.NewFileStore(mx.DataDir)
	//Email configurations and sync
	email.DB = dconf.Conn
	email.LOG = lge
	econf.MxClient = mx.Client
	email.StartMailSchedeuler(mx.Client)

	go func() {
		<-q
		clear(done)
	}()
	<-done
}

func clear(done chan struct{}) {
	defer close(done)
	email.Close()
}
