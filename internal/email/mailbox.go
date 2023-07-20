package email

import (
	"fmt"

	"gorm.io/gorm"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

func Mailbox(client *mautrix.Client, roomID id.RoomID) error {
	imapAccount, err := GetIMAPAccount(string(roomID))
	if err == gorm.ErrRecordNotFound {
		client.SendText(roomID, "You have to login a email account, Use !setup to login IMAP account!")
		return err
	}
	if err != nil {
		return err
	}

	client.SendText(roomID, fmt.Sprintf("Current IMAP account:  %s", imapAccount.UserName))
	return nil
}
