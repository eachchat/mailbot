package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
)

// New Database connection
func (d *DBCONF) New() (err error) {
	switch d.Type {
	case "sqlite":
		d.Conn, err = gorm.Open(sqlite.Open(d.DBName), &gorm.Config{})
	case "postgresql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", d.Host, d.UserName, d.Password, d.DBName, d.Port)
		d.Conn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", d.UserName, d.Password, d.Host, d.Port, d.DBName)
		d.Conn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	}
	return err
}

// InitDB  init database
func (d *DBCONF) InitDB() error {
	var mail Mails
	err := d.Conn.AutoMigrate(mail)
	if err != nil {
		return err
	}
	var rooms Rooms
	err = d.Conn.AutoMigrate(rooms)
	if err != nil {
		return err
	}
	var imapA ImapAccounts
	err = d.Conn.AutoMigrate(imapA)
	if err != nil {
		return err
	}
	var smtpA SmtpAccounts
	err = d.Conn.AutoMigrate(smtpA)
	if err != nil {
		return err
	}
	var recent Recentevent
	err = d.Conn.AutoMigrate(recent)
	if err != nil {
		return err
	}

	return nil
}
