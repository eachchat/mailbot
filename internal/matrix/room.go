package matrix

import (
	"fmt"

	"github.com/eachchat/mailbot/internal/email"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

/*
	func getRoomInfo(roomID string) (string, error) {
		var imapA db.ImapAccounts
		var smtpA db.SmtpAccounts
		infoText := ""
		tx := DB.Where(&db.ImapAccounts{RoomID: roomID}).First(&imapA)
		if tx.Error == gorm.ErrRecordNotFound {
			infoText += "IMAP account: None"
		} else {
			infoText += fmt.Sprintf("IMAP account: None", imapA.UserName)
		}
		tx = DB.Where(&db.SmtpAccounts{RoomID: roomID}).First(&smtpA)
		if tx.Error == gorm.ErrRecordNotFound {
			infoText += "SMPT account: None"
		} else {
			infoText += fmt.Sprintf("SMTP account: None", smtpA.UserName)
		}
		return infoText, nil
	}

	func getRoomPKID(roomID string) (int, error) {
		var r db.Rooms
		tx := DB.Where(&db.Rooms{RoomID: roomID}).First(&r)
		if tx.Error != nil {
			return -1, tx.Error
		}
		return r.PkID, nil
	}

	func setHTMLenabled(roomID string, enabled bool) error {
		var isenabled int
		if enabled {
			isenabled = 1
		}
		tx := DB.Model(&db.Rooms{}).Where("room_id = ?", roomID).Update("is_html_enabled", isenabled)
		return tx.Error
	}
*/
func (mx *MxConf) ViewMailboxes(roomID string, client *mautrix.Client) {
	var ign bool
	imapA, _ := email.GetRoomAccounts(roomID)
	if imapA == nil {
		mx.Client.SendText(id.RoomID(roomID), "You have to setup an IMAP account to use this command. Use !setup for more informations")
		//mx.Client.SendText(id.RoomID(roomID), "群聊暂未登录邮箱。使用!setup获取更多详情")
		return
	} else {
		if imapA.IgnoreSSL == 1 {
			ign = true
		}
		c, err := email.LoginMail(imapA.Host, imapA.UserName, imapA.Password, ign)
		if err != nil {
			LOG.Error().Msg(fmt.Sprintf("Failed to login IMAP with user: %s", imapA.UserName))
			mx.Client.SendText(id.RoomID(roomID), fmt.Sprintf("Can not to login IMAP server, user: %s", imapA.UserName))
		}
		mailboxes, err := getMailboxes(c)
		if err != nil {
			LOG.Error().Msg("#47 getMailboxes: " + err.Error())
			mx.Client.SendText(id.RoomID(roomID), "An server-error occured Errorcode: #47")
			return
		}
		mx.Client.SendText(id.RoomID(roomID), "Your mailboxes:\r\n"+mailboxes+"\r\nUse !setmailbox <mailbox> to change your mailbox")
		//mx.Client.SendText(id.RoomID(roomID), "您的邮件接收箱列表：\r\n\r\n"+mailboxes+"\r\n发送 !mail set mailbox [mailbox name] 以变更邮件接收箱的推送。")
	}
}

func getMailboxes(emailClient *client.Client) (string, error) {
	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 20)
	mboxes := ""
	if err := emailClient.List("", "*", mailboxes); err != nil {
		LOG.Error().Msg(err.Error())
		return "", err
	}
	for m := range mailboxes {
		mboxes += "-> " + m.Name + "\r\n"
	}
	return mboxes, nil
}
