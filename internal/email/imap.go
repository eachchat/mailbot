package email

import (
	"fmt"
	"strings"
	"time"

	"github.com/eachchat/mailbot/internal/db"
	"github.com/eachchat/mailbot/pkg/utils"
	"gorm.io/gorm"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

func SetupImap(roomID id.RoomID, msg []string, sender id.UserID, mailCheckInterval int, isEnableHTML bool, mxClient *mautrix.Client) error {
	var ign int
	if len(msg) < 5 {
		return fmt.Errorf("wrong command: !setup imap,imap.example.com:993,user@exapmole.com,PASSWORD,INBOX,ture/false")
	}
	if isImapAccountAlreadyInUse(roomID.String()) {
		return fmt.Errorf("IMAP has already set up, the mail: %s", msg[1])
	}
	if strings.TrimSpace(strings.ToLower(msg[4])) == "true" {
		ign = 1
	}
	isok, err := hasRoom(roomID.String())
	if err == gorm.ErrRecordNotFound || !isok {
		insertNewRoom(roomID.String(), isEnableHTML)
	}
	insertImapAccount(roomID.String(), msg[0], msg[1], utils.B64Encode(msg[2]), msg[3], mailCheckInterval, ign, sender.String())
	errcount := 0
	for {
		count, err := waitForMailboxReady(string(roomID), msg[3])
		if err != nil {
			LOG.Error().Msg(fmt.Sprintf("Waiting for  imap  account ready  retry time %d, error: %s", errcount, err.Error()))
			if errcount > 2 {
				time.Sleep(1 * time.Second)
				continue
			}
			time.Sleep(1 * time.Second)
			errcount++
			continue
		}
		if count < 1 {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	imapAcc, _ := GetRoomAccounts(roomID.String())
	imapAcc.Silence = true
	go StartMailListener(*imapAcc, mxClient)
	return nil
}

func insertImapAccount(roomID, host, username, password, mailbox string, mailCheckInterval, ignoreSSl int, set_by string) error {
	tx := DB.Create(&db.ImapAccounts{RoomID: roomID, Host: host, UserName: username,
		Password: password, Mailbox: mailbox, IgnoreSSL: ignoreSSl, SetBy: set_by, MailCheckInterval: mailCheckInterval})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func isImapAccountAlreadyInUse(email string) bool {
	var count int64
	DB.Model(&db.ImapAccounts{}).Where("user_name = ?", email).Count(&count)
	return count >= 1
}

func getimapAccounts() ([]db.ImapAccounts, error) {
	var list []db.ImapAccounts
	tx := DB.Model(&db.ImapAccounts{}).Find(&list)
	if tx.Error != nil {
		return list, tx.Error
	}
	return list, nil
}

func GetIMAPAccount(roomID string) (*db.ImapAccounts, error) {
	var imp db.ImapAccounts
	tx := DB.Model(&db.ImapAccounts{}).Where("room_id = ?", roomID).First(&imp)
	if tx.Error != nil {
		return nil, tx.Error
	}
	//pass, berr := base64.StdEncoding.DecodeString(password)

	return &imp, nil
}

/*
func saveMailbox(roomID, newMailbox string) error {
	tx := DB.Model(&db.ImapAccounts{}).Where("room_id = ?", roomID).Update("mailbox", newMailbox)
	return tx.Error
}


func getMailbox(roomID string) (string, error) {
	var imapA db.ImapAccounts
	tx := DB.Where(&db.ImapAccounts{RoomID: roomID}).First(&imapA)
	return imapA.Mailbox, tx.Error
}
*/

func insertNewRoom(roomID string, isEnableHTML bool) error {
	var r = db.Rooms{RoomID: roomID, IsHTMLenabled: isEnableHTML}
	tx := DB.Create(&r)
	return tx.Error
}

func hasRoom(roomID string) (bool, error) {
	var count int64
	tx := DB.Model(&db.Rooms{}).Where("room_id = ?", roomID).Count(&count)
	if tx.Error != nil {
		return false, tx.Error
	}
	return (count >= 1), nil
}

func waitForMailboxReady(roomID, mbox string) (int64, error) {
	var count int64
	tx := DB.Model(&db.ImapAccounts{}).Where("room_id = ? and mailbox = ?", roomID, mbox).Count(&count)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return count, nil
}
