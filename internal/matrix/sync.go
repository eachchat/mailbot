package matrix

import (
	"fmt"
	"strings"

	"github.com/eachchat/mailbot/internal/email"
	"gorm.io/gorm"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (mx *MxConf) syncMemberJoin(source mautrix.EventSource, evt *event.Event) {
	LOG.Info().Msg(fmt.Sprintf("You have joined into the room %s", evt.RoomID.String()))
	host := getHostFromMatrixID(string(evt.Sender))
	LOG.Debug().Msg(fmt.Sprintf("allow servers: %v, user host: %v", mx.AllowedServers, host))
	if contains(mx.AllowedServers, host) {
		mx.Client.JoinRoom(string(evt.RoomID), "", nil)
		mx.Client.SendNotice(evt.RoomID, "Hey you have invited me to a new room. Enter !help for command informations")
		//mx.Client.SendNotice(evt.RoomID, "您好！欢迎使用邮箱助手。\r\n请使用!login配置您的邮箱账户。")
	} else {
		//mx.Client.SendNotice(evt.RoomID, "您不被允许添加邮件助手。")
		mx.Client.SendNotice(evt.RoomID, "You are not allowed to join the room")
		mx.Client.LeaveRoom(evt.RoomID)
	}
}

func (mx *MxConf) syncMemberShip(source mautrix.EventSource, evt *event.Event) {
	evtType := evt.Content.AsMember().Membership
	LOG.Debug().Msg(fmt.Sprintf("Event type: %v, event userID: %s", evtType, evt.Sender.String()))
	if evtType == event.MembershipBan || evtType == event.MembershipLeave && evt.Sender != mx.Client.UserID {
		logOut(mx.Client, evt.RoomID.String(), true)
	}
}

func (mx *MxConf) syncMessageEvnt(source mautrix.EventSource, evt *event.Event) {
	LOG.Debug().Msg(evt.Content.AsMessage().Body)
	if evt.Sender == mx.Client.UserID {
		return
	}

	ts, err := getRecentEventTs(evt.RoomID.String(), string(event.EventMessage.Type))
	switch err {
	case gorm.ErrRecordNotFound:
		err = insertRecentEvent(string(evt.RoomID), string(event.EventMessage.Type), evt.Timestamp)
		if err != nil {
			LOG.Error().Msg(fmt.Sprintf("saveRecentEvent: %s", err.Error()))
		}
	case nil:
		if ts >= evt.Timestamp {
			return
		} else {
			err = saveRecentEvent(string(evt.RoomID), string(event.EventMessage.Type), evt.Timestamp)
			if err != nil {
				LOG.Error().Msg(fmt.Sprintf("saveRecentEvent: %s", err.Error()))
			}
		}
	default:
		LOG.Error().Msg(fmt.Sprintf("getRecentEventTs: %s", err.Error()))
	}
	message := evt.Content.AsMessage().Body
	roomID := evt.RoomID
	if message == "" {
		return
	}
	switch true {
	case strings.HasPrefix(message, "!send"):
		account, err := email.GetSMTPAccount(roomID.String())
		if err == gorm.ErrRecordNotFound {
			mx.Client.SendText(roomID, "You have to set up a SMTP account, use !setup for more information")
			return
		}
		err = email.SendMailMsg(message, roomID.String(), account)
		if err != nil {
			mx.Client.SendText(roomID, err.Error())
		} else {
			mx.Client.SendText(roomID, "Send email successfully.")
		}
		return
	case strings.HasPrefix(message, "!login"):
		mx.login(roomID)
	case strings.HasPrefix(message, "!setup"):
		message = strings.TrimSpace(strings.Replace(message, "!setup", "", 1))
		msgSlice := strings.Split(message, ",")
		switch strings.TrimSpace(msgSlice[0]) {
		case "smtp":
			err := email.SetupSmtp(msgSlice[1:], roomID)
			if err != nil {
				mx.Client.SendNotice(evt.RoomID, err.Error())
			} else {
				mx.Client.SendNotice(evt.RoomID, "SMTP has been set up successfully!")
			}
		case "imap":
			err := email.SetupImap(roomID, msgSlice[1:], evt.Sender, mx.DefaultMailCheckInterval, mx.HtmlDefault, mx.Client)
			if err != nil {
				mx.Client.SendNotice(evt.RoomID, err.Error())
			} else {
				mx.Client.SendNotice(evt.RoomID, "IMAP has been set up successfully!")
			}
		default:
			mx.Client.SendNotice(roomID, "!setup imap/smtp, host:port, username(em@ail.com),password,<mailbox (only for imap)>,ignoreSSLcert(true/false)")
		}
		/*
			case strings.HasPrefix(message, "!setmailbox"):
				message = strings.TrimSpace(strings.Replace(message, "!setmailbox", "", 1))
				if message == "" {
					mx.Client.SendText(roomID, "Usage: !setmailbox <new mailbox>")
				}
				email.SetMailbox(mx.Client, roomID, message)
				//email.ViewMailboxes(roomID.String(), client)
				return
		*/
	case strings.HasPrefix(message, "!mailboxes"):
		mx.ViewMailboxes(roomID.String(), mx.Client)
	case strings.HasPrefix(message, "!logout") || strings.HasPrefix(message, "!leave"):
		err = logOut(mx.Client, roomID.String(), false)
		if err != nil {
			mx.Client.SendText(roomID, "Error logging out: "+err.Error())
		} else {
			//client.SendText(roomID, "您已移除了邮箱"+imapAccount.username+"。")
			mx.Client.SendText(roomID, fmt.Sprintf("Email account has removed from account %s.", evt.Sender.String()))
		}
	case strings.HasPrefix(message, "!mailbox"):
		email.Mailbox(mx.Client, id.RoomID(roomID.String()))
	default:
		mx.ReturnHelp(roomID)
	}

}

func (mx *MxConf) startMatrixSync() {
	LOG.Info().Msg(fmt.Sprintf("Starting to listen room message, userID: %s", mx.Client.UserID))

	syncer := mx.Client.Syncer.(*mautrix.DefaultSyncer)

	// Join room
	//syncer.OnEventType(event.StateJoinRules, mx.syncMemberJoin)
	syncer.OnEventType(event.StateJoinRules, mx.syncMemberJoin)

	// Leave room action
	syncer.OnEventType(event.StateMember, mx.syncMemberShip)

	syncer.OnEventType(event.EventMessage, mx.syncMessageEvnt)

	err := mx.Client.Sync()
	if err != nil {
		LOG.Error().Msg(fmt.Sprintf("Syncing error: %s", err.Error()))
	}

}
