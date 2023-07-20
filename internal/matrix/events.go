package matrix

import "github.com/eachchat/mailbot/internal/db"

func getRecentEventTs(roomID, eventType string) (ts int64, err error) {
	var rcet db.Recentevent
	tx := DB.Model(&db.Recentevent{}).Where("room_id = ? and event_type = ?", roomID, eventType).First(&rcet)
	ts = rcet.Ts
	err = tx.Error
	return
}

func insertRecentEvent(roomID, eventType string, ts int64) (err error) {
	tx := DB.Model(&db.Recentevent{Ts: ts, RoomID: roomID, EventType: eventType})
	err = tx.Error
	return
}

func saveRecentEvent(roomID, eventType string, ts int64) (err error) {
	tx := DB.Model(&db.Recentevent{}).Where("room_id = ? and event_type = ?", roomID, eventType).Update("ts", ts)
	err = tx.Error
	return
}
