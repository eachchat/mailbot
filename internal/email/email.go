package email

import (
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/eachchat/mailbot/internal/db"
	"github.com/eachchat/mailbot/pkg/utils"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type EmailConf struct {
	MxClient *mautrix.Client
	Logger   *zerolog.Logger
}

const maxRoomChecks = 15
const maxErrUntilReconnect = 10

var DB *gorm.DB
var LOG *zerolog.Logger
var clients map[string]*client.Client
var imapErrors map[string]*imapError
var checksPerAccount map[string]int
var listenerMap map[string]chan bool

type imapError struct {
	retryCount, loginErrCount int
}

func insertEmail(email string, roomID string) error {
	tx := DB.Create(&db.Mails{RoomID: roomID, Mail: email})
	return tx.Error
}

func dbContainsMail(mail string, roomID string) (bool, error) {
	var count int64

	tx := DB.Model(&db.Mails{}).Where("mail = ? and room_id = ?", mail, roomID).Count(&count)
	if tx.Error != nil {
		return false, tx.Error
	}
	return (count > 0), nil
}

// DeleteMails delete mails of the room
func DeleteMails(roomID string) {
	DB.Model(&db.Mails{}).Where("room_id = ?", roomID).Delete(&db.Mails{})
}

func SyncMail(roomID id.RoomID, imapA db.ImapAccounts, mxclient *mautrix.Client) {
	var ignoreSSlCert bool
	if imapA.IgnoreSSL == 1 {
		ignoreSSlCert = true
	} else {
		ignoreSSlCert = false
	}
	mclient, err := LoginMail(imapA.Host, imapA.UserName, imapA.Password, ignoreSSlCert)
	if mclient == nil || err != nil {
		mxclient.SendText(roomID, "Can not create email bridge! Error: "+err.Error())
		return
	}
	//mxclient.SendText(roomID, "您已成功登录了邮箱 "+username+"，您可以在亿洽内同步接收新邮件")
	mxclient.SendText(roomID, "You have already login IMAP with "+imapA.UserName)
	StartMailListener(imapA, mxclient)
}

func StartMailListener(account db.ImapAccounts, mxClient *mautrix.Client) {

	var ign bool
	if account.IgnoreSSL == 1 {
		ign = true
	}
	quit := make(chan bool)

	for {
		var mClient *client.Client
		mClient, err := LoginMail(account.Host, account.UserName, account.Password, ign)
		LOG.Debug().Msg(fmt.Sprintf("StartMailListener, user: %s, host: %s, password: (%s)", account.UserName, account.Host, utils.B64Decode(account.Password)))
		if err != nil {
			LOG.Error().Msg(fmt.Sprintf("Failed to login IMAP, StartMailListener: %s", err.Error()))
			time.Sleep(1 * time.Minute)
			continue
		}
		clients[account.RoomID] = mClient
		listenerMap[account.RoomID] = quit
		break

	}

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				if getChecksForAccount(account.RoomID) >= maxRoomChecks {
					reconnect(account, mxClient)
					return
				}
				fetchNewMails(clients[account.RoomID], mxClient, &account)
				checksPerAccount[account.RoomID]++
				time.Sleep((time.Duration)(account.MailCheckInterval) * time.Second)
			}
		}
	}()
}

func getChecksForAccount(roomID string) int {
	checks, ok := checksPerAccount[roomID]
	if ok {
		return checks
	}
	checksPerAccount[roomID] = 0
	return 0
}

func reconnect(account db.ImapAccounts, mxClient *mautrix.Client) {
	checksPerAccount[account.RoomID] = 0
	StopMailChecker(account.RoomID)
	nacc := account
	go StartMailListener(nacc, mxClient)
}

func StopMailChecker(roomID string) {
	_, ok := listenerMap[roomID]
	if ok {
		close(listenerMap[roomID])
		delete(listenerMap, roomID)
	}
}

func fetchNewMails(mClient *client.Client, mxClient *mautrix.Client, account *db.ImapAccounts) {

	LOG.Debug().Msg("starting to fetch new email.....")
	messages := make(chan *imap.Message, 1)
	section, errCode := getMails(mClient, account.Mailbox, messages)
	if section == nil {
		if errCode == 0 {
			haserr, errCount := hasError(account.RoomID)
			if haserr {
				if imapErrors[account.RoomID].loginErrCount > 15 {
					mxClient.SendNotice(id.RoomID(account.RoomID), "You have got too much errors for the emailaccount: "+account.UserName)
				}
				if errCount < maxErrUntilReconnect {
					imapErrors[account.RoomID].retryCount++
				} else {
					imapErrors[account.RoomID].retryCount = 0
					imapErrors[account.RoomID].loginErrCount++
					reconnect(*account, mxClient)
					return
				}
			}
		}
		if account.Silence {
			account.Silence = false
		}
		return
	}

	for msg := range messages {
		mailID := msg.Envelope.Subject + strconv.Itoa(int(msg.InternalDate.Unix()))
		if has, err := dbContainsMail(mailID, account.RoomID); !has && err == nil {
			go insertEmail(mailID, account.RoomID)
			if !account.Silence {
				handleMail(msg, section, mxClient, *account)
			}
		} else if err != nil {
			fmt.Println(err.Error())
		}
	}
	if account.Silence {
		account.Silence = false
	}
}

func handleMail(mail *imap.Message, section *imap.BodySectionName, mxClient *mautrix.Client, account db.ImapAccounts) {
	content := getMailContent(mail, section, account.RoomID)
	if content == nil {
		return
	}
	/*
		for _, senderMail := range content.sendermails {

			if checkForBlocklist(account.RoomID, senderMail) {
				return
			}
		}
	*/
	from := html.EscapeString(content.from)

	headerContent := &event.MessageEventContent{
		Format: event.FormatHTML,
		//Body:   "\r\n────────────────────────────────────\r\n## You've got a new Email from " + from + "\r\n" + "Subject: " + content.subject + "\r\n" + "────────────────────────────────────",
		Body: fmt.Sprintf("\r\n────────────────────────────────────\r\n发件人:  %s\r\n主题: %sr\n────────────────────────────────────", from, content.subject),
		//FormattedBody: " You've got a new Email</b> from <b>" + from + "</b><br>" + "Subject: " + content.subject + "<br>" + "────────────────────────────────────",
		FormattedBody: fmt.Sprintf("<br>────────────────────────────────────<br><b>发件人: </b> %s<br> <b>主题: </b> %s <br>────────────────────────────────────", from, content.subject),
		MsgType:       event.MsgText,
	}

	mxClient.SendMessageEvent(id.RoomID(account.RoomID), event.EventMessage, &headerContent)
	for _, v := range content.Mbody {
		switch v.format {
		case "text/html":
			bodyContent := &event.MessageEventContent{
				Format: event.FormatHTML,
				Body:   string(v.body),
				//FormattedBody: string(markdown.ToHTML(v.body, nil, nil)),
				FormattedBody: strings.ReplaceAll(string(v.body), "\r\n", ""),
				MsgType:       event.MsgText,
			}
			mxClient.SendMessageEvent(id.RoomID(account.RoomID), event.EventMessage, &bodyContent)
		case "text/plain":
			mxClient.SendText(id.RoomID(account.RoomID), string(v.body))
		default:
			mdResp, err := uploadMedia(mxClient, v.body, v.format)
			if err != nil {
				LOG.Error().Msg(fmt.Sprintf("Failed to upload media, mimetype: %s filename: %s", v.format, v.fileName))
				return
			}
			bodyContent := &event.MessageEventContent{
				MsgType: event.MsgImage,
				Body:    v.fileName,
				Info: &event.FileInfo{
					MimeType: v.format,
				},
				URL: id.ContentURIString(mdResp.ContentURI.String()),
			}
			mxClient.SendMessageEvent(id.RoomID(account.RoomID), event.EventMessage, &bodyContent)
		}
	}
	for _, v := range content.Attach {
		mdResp, err := uploadMedia(mxClient, v.attachmentCnt, v.attachment)
		if err != nil {
			LOG.Error().Msg(fmt.Sprintf("Failed to upload attachement media, mimetype: %s filename: %s", v.format, v.attachment))
			return
		}
		bodyContent := &event.MessageEventContent{
			MsgType:  event.MsgFile,
			Body:     v.attachment,
			FileName: v.attachment,
			Info: &event.FileInfo{
				MimeType: v.format,
			},
			URL: id.ContentURIString(mdResp.ContentURI.String()),
		}

		mxClient.SendMessageEvent(id.RoomID(account.RoomID), event.EventMessage, &bodyContent)
		//mxClient.SendText(id.RoomID(account.RoomID), fmt.Sprintf("附件名：%v, 大小：%v", v.attachment, v.attachmentSize))
	}
	tailContent := &event.MessageEventContent{
		Format: event.FormatHTML,
		//Body:   "\r\n────────────────────────────────────\r\n## You've got a new Email from " + from + "\r\n" + "Subject: " + content.subject + "\r\n" + "────────────────────────────────────",
		Body: fmt.Sprintf("\r\n────────────────主题:【%s】邮件结束────────────────────", content.subject),
		//FormattedBody: " You've got a new Email</b> from <b>" + from + "</b><br>" + "Subject: " + content.subject + "<br>" + "────────────────────────────────────",
		FormattedBody: fmt.Sprintf("<br>────────────────<b>主题:【%s】邮件结束 </b>────────────────────", content.subject),
		MsgType:       event.MsgText,
	}
	mxClient.SendMessageEvent(id.RoomID(account.RoomID), event.EventMessage, &tailContent)
}

func StartMailSchedeuler(mxClient *mautrix.Client) {
	listenerMap = make(map[string]chan bool)
	clients = make(map[string]*client.Client)
	imapErrors = make(map[string]*imapError)
	checksPerAccount = make(map[string]int)
	LOG.Info().Msg("Starting to listen email....")
	accounts, err := getimapAccounts()
	if err != nil {
		log.Panic(err)
	}
	for _, v := range accounts {
		LOG.Debug().Msg(fmt.Sprintf("user: %s, host: %s", v.UserName, v.Host))
		go StartMailListener(v, mxClient)
	}
}

func hasError(roomID string) (has bool, count int) {
	_, ok := imapErrors[roomID]
	if ok {
		return true, imapErrors[roomID].retryCount
	}
	return false, -1
}

func uploadMedia(mclient *mautrix.Client, body []byte, format string) (*mautrix.RespMediaUpload, error) {
	mediaID, _ := mclient.CreateMXC()

	mediaResp, err := mclient.UploadMedia(mautrix.ReqUploadMedia{
		ContentBytes:      body,
		ContentType:       format,
		MXC:               mediaID.ContentURI,
		UnstableUploadURL: mediaID.UnstableUploadURL,
	})
	if err != nil {
		LOG.Error().Msg(fmt.Sprintf("failed to upload media with err: %s", err.Error()))
		return nil, err
	}
	return mediaResp, err
}
