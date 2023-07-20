package email

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"maunium.net/go/mautrix/id"

	"github.com/eachchat/mailbot/internal/db"
	"github.com/eachchat/mailbot/pkg/utils"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"maunium.net/go/mautrix"
)

type emailContent struct {
	from, to, subject string
	sendermails       []string
	date              time.Time
	Mbody             []emailBody
	Attach            []attachments
}

type emailBody struct {
	body     []byte
	format   string
	fileName string
}

type attachments struct {
	format        string
	attachment    string
	attachmentCnt []byte
}

func LoginMail(host, username, password string, ignoreSSL bool) (*client.Client, error) {
	ailClient, err := client.DialTLS(host, &tls.Config{InsecureSkipVerify: ignoreSSL})
	if err != nil {
		return nil, err
	}

	if err := ailClient.Login(username, strings.TrimSpace(utils.B64Decode(password))); err != nil {
		return nil, err
	}
	return ailClient, nil
}

func getMails(mClient *client.Client, mBox string, messages chan *imap.Message) (*imap.BodySectionName, int) {

	mbox, err := mClient.Select(mBox, false)
	if err != nil {
		LOG.Error().Msg("#12 couldnt get " + mBox + ", " + err.Error())
		return nil, 0
	}

	if mbox == nil {
		LOG.Error().Msg("#23 getMails mbox is nil")
		return nil, 0
	}
	if mbox.Messages == 0 {
		LOG.Error().Msg("#13 getMails no messages in inbox ")
		return nil, 1
	}

	seqSet := new(imap.SeqSet)
	maxMessages := uint32(5)
	if mbox.Messages < maxMessages {
		maxMessages = mbox.Messages
	}

	for i := uint32(0); i < maxMessages; i++ {
		seqSet.AddNum(mbox.Messages - i)
	}

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()}
	go func() {
		if err := mClient.Fetch(seqSet, items, messages); err != nil {
			LOG.Error().Msg("#14 couldnt fetch messages: " + err.Error())
		}
	}()
	return section, -1
}

func getMailContent(msg *imap.Message, section *imap.BodySectionName, roomID string) *emailContent {
	if msg == nil {
		fmt.Println("msg is nil")
		LOG.Error().Msg("#15 getMailContent msg is nil")
		return nil
	}

	r := msg.GetBody(section)
	if r == nil {
		fmt.Println("reader is nli")
		LOG.Error().Msg("#16 getMailContent r (reader) is nil")
		return nil
	}

	jmail := emailContent{}
	mr, err := mail.CreateReader(r)
	if err != nil {
		fmt.Println(err.Error())
		LOG.Error().Msg("#17 getMailContent create reader err: " + err.Error())
		return nil
	}

	header := mr.Header
	if date, err := header.Date(); err == nil {
		log.Println("Date:", date)
		jmail.date = date
	}
	if from, err := header.AddressList("From"); err == nil {
		list := make([]string, len(from))
		for i, sender := range from {
			if len(sender.Name) > 0 {
				list[i] = sender.Name + "<" + sender.Address + ">"
			} else {
				list[i] = sender.Address
			}
			jmail.sendermails = append(jmail.sendermails, sender.Address)
		}
		jmail.from = strings.Join(list, ",")
	}
	if to, err := header.AddressList("To"); err == nil {
		list := make([]string, len(to))
		for i, receiver := range to {
			if len(receiver.Name) > 0 {
				list[i] = receiver.Name + "<" + receiver.Address + ">"
			} else {
				list[i] = receiver.Address
			}
		}
		jmail.to = strings.Join(list, ",")
	}
	if subject, err := header.Subject(); err == nil {
		log.Println("Subject:", subject)
		jmail.subject = subject
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(err.Error())
			LOG.Error().Msg("#18 getMailContent nextPart err: " + err.Error())
			break
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			var eb emailBody
			b, _ := io.ReadAll(p.Body)
			eb.body = b
			mimety := http.DetectContentType(b[:512])
			if mimety == "application/octet-stream" {
				eb.format = "text/plain"
			} else {
				eb.format = strings.Split(mimety, ";")[0]
			}
			eb.fileName = utils.RandomStr(10) + utils.GetFileSubfix(mimety)
			jmail.Mbody = append(jmail.Mbody, eb)
		case *mail.AttachmentHeader:
			var att attachments
			att.attachment, _ = h.Filename()
			b, _ := io.ReadAll(p.Body)
			att.attachmentCnt = b
			mimety := http.DetectContentType(b[:512])
			att.format = mimety
			jmail.Attach = append(jmail.Attach, att)
		}
	}
	return &jmail
}

/*
	func parseMailBody(body *string) {
		*body = strings.ReplaceAll(*body, "<br>", "\r\n")
		*body = strip.StripTags(html.UnescapeString(*body))
	}
*/
func SetMailbox(client *mautrix.Client, roomID id.RoomID, mailbox string) {
	imapAcc, _ := GetRoomAccounts(roomID.String())
	if imapAcc != nil {
		saveMailbox(roomID.String(), mailbox)
		DeleteMails(roomID.String())
		StopMailChecker(roomID.String())
		imapAcc.Silence = true
		go StartMailListener(*imapAcc, client)
		client.SendText(roomID, "Mailbox updated")

	} else {
		client.SendText(roomID, "You have to setup an IMAP account to use this command. Use !setup or !login for more informations")
	}
}

/*
	func isHTMLenabled(roomID string) (bool, error) {
		var r db.Rooms
		tx := DB.Model(&db.Rooms{}).Where("room_id = ?", roomID).First(&r)
		if tx.Error != nil {
			return false, tx.Error
		}
		if r.IsHTMLenabled {
			return true, nil
		} else {
			return false, nil
		}
	}
*/
func GetRoomAccounts(roomID string) (*db.ImapAccounts, *db.SmtpAccounts) {
	var imapAcc db.ImapAccounts
	var smtpACC db.SmtpAccounts
	DB.Model(&db.SmtpAccounts{}).Where("room_id = ?", roomID).First(&smtpACC)
	DB.Model(&db.ImapAccounts{}).Where("room_id = ?", roomID).First(&imapAcc)

	return &imapAcc, &smtpACC
}
