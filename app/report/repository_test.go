package report

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestGetUserMessagesStat(t *testing.T) {
	db := sqlx.MustConnect("sqlite", ":memory:")
	db.MustExec(`create table "user"(tg_chat_id integer, tg_username text)`)
	db.MustExec(`create table message(chat_id integer, role text, created_at integer)`)
	db.MustExec(`insert into "user" values (1, 'alice')`)
	now := time.Now().Unix()
	db.MustExec(`insert into message values (1, 'user', ?), (1, 'user', ?), (1, 'assistant', ?), (1, 'user', ?)`,
		now, now, now, now-7200)

	rep := NewDbReportRepository(db)
	stats, err := rep.GetUserMessagesStat(time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 || stats[0].TgName != "alice" || stats[0].MessageQt != 2 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}
