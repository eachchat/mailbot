package matrix

import (
	"fmt"
	"strings"

	"github.com/eachchat/mailbot/internal/db"
	"github.com/eachchat/mailbot/internal/email"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

/*
	func getHostFromMatrixID(matrixID string) (host string, err int) {
		if strings.Contains(matrixID, ":") {
			splt := strings.Split(matrixID, ":")
			if len(splt) == 2 {
				return splt[1], -1
			}
			return "", 1
		}
		return "", 0
	}
*/
func contains(a []string, x string) bool {
	for _, n := range a {
		fmt.Println(x, n)
		if x == n {
			return true
		}
	}
	return false
}

func logOut(client *mautrix.Client, roomID string, leave bool) error {
	email.StopMailChecker(roomID)
	deleteRoomAndEmailByRoomID(roomID)
	if leave {
		_, err := client.LeaveRoom(id.RoomID(roomID))
		if err != nil {
			LOG.Error().Msg("#65 bot can't leave room: " + err.Error())
			return err
		}
	}
	return nil
}

func deleteRoomAndEmailByRoomID(roomID string) {
	DB.Model(&db.ImapAccounts{}).Where("room_id = ?", roomID).Delete(&db.ImapAccounts{})

	DB.Model(&db.SmtpAccounts{}).Where("room_id = ?", roomID).Delete(&db.SmtpAccounts{})

	DB.Model(&db.Rooms{}).Where("room_id = ?", roomID).Delete(&db.Rooms{})
}

func (mx *MxConf) login(roomID id.RoomID) {
	msg := ""
	imapA, smtpA := email.GetRoomAccounts(roomID.String())
	if smtpA != nil {
		msg += fmt.Sprintf("SMTP:\n\tHost: %v\n\tPort: %v\n\tUser: %v\n", smtpA.Host, smtpA.Port, smtpA.UserName)
	}
	if imapA != nil {
		msg += fmt.Sprintf("IMAP:\n\tHost: %v\n\tUser: %v", imapA.Host, imapA.UserName)
	}
	mx.Client.SendNotice(roomID, msg)
}

func (mx *MxConf) ReturnHelp(roomID id.RoomID) {
	//WriteLog(info, "roomID: "+string(roomID)+", sender: "+string(sender))
	helpText := "-------- Help --------\r\n"
	helpText += "!setup imap/smtp,host:port,username(user@example.com),password,<mailbox (only for imap)>,ignoreSSLcert(true/false)\r\n"
	//helpText += "!ping - gets information about the email bridge for this room\r\n"
	helpText += "!help - shows this command help overview\r\n"
	helpText += "!mailboxes - shows a list with all mailboxes available on your IMAP server\r\n"
	helpText += "!setmailbox (mailbox) - changes the mailbox for the room\r\n"
	helpText += "!mailbox - shows the currently selected mailbox\r\n"
	//helpText += "!sethtml (on/off or true/false) - sets HTML-rendering for messages on/off\r\n"
	helpText += "!logout remove email bridge from current room\r\n"
	helpText += "!leave unbridge the current room and kick the bot\r\n"
	//helpText += "\r\n---- Email writing commands ----\r\n"
	helpText += "!send - sends the email\r\n"
	//helpText += "!rm <file> - removes given attachment from email\r\n"

	mx.Client.SendText(roomID, helpText)
}

func getHostFromMatrixID(matrixID string) (host string) {
	if strings.Contains(matrixID, ":") {
		splt := strings.Split(matrixID, ":")
		if len(splt) == 2 {
			return splt[1]
		}
		return ""
	}
	return ""
}

