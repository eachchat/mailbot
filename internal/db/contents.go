package db

import (
	"gorm.io/gorm"
)

// DBCONF the configurations for db connection
type DBCONF struct {
	Type     string   `yaml:"type"`
	UserName string   `yaml:"username"`
	Password string   `yaml:"password"`
	DBName   string   `yaml:"dbName"`
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	Conn     *gorm.DB `yaml:"-"`
}

// Mail the table for mail
type Mails struct {
	Mail   string
	RoomID string
}

// Rooms table name rooms
type Rooms struct {
	//PkID          int `gorm:",autoIncrement,primaryKey"`
	RoomID        string `gorm:"primaryKey"`
	IsHTMLenabled bool   `gorm:"default:0"`
}

// ImapAccounts table of imap account info
type ImapAccounts struct {
	//PkID              int `gorm:"AUTO_INCREMENT,primaryKey"`
	Host              string
	UserName          string
	Password          string
	IgnoreSSL         int
	Mailbox           string
	SetBy             string
	RoomID            string `gorm:"primaryKey"`
	MailCheckInterval int
	Silence           bool
}

// SmtpAccounts table of imap account info
type SmtpAccounts struct {
	//PkID      int `gorm:",AUTO_INCREMENT,primaryKey"`
	Host      string
	UserName  string
	Password  string
	IgnoreSSL int
	Port      int
	RoomID    string `gorm:"primaryKey"`
}

// Blocklist table of blocklist
type Blocklist struct {
	PkID        int `gorm:"primaryKey,AUTO_INCREMENT"`
	ImapAccount int
	Address     int
}

// Recentevent  table of envent
type Recentevent struct {
	PkID      int `gorm:"AUTO_INCREMENT,primaryKey"`
	Ts        int64
	EventType string
	RoomID    string
}
