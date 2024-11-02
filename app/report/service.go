package report

import (
	"amArbaoui/yaggptbot/app/storage"
	"amArbaoui/yaggptbot/app/telegram"
	"amArbaoui/yaggptbot/app/util"
	"bytes"
	"log"
	"text/template"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

type TgReportRepository interface {
	GetUserLastMessage() ([]*UserLastMessage, error)
	GetUserMessagesStat(after time.Time) ([]*UserMessageStat, error)
}

type ReportService struct {
	rep     TgReportRepository
	chatSrv telegram.ChatService
	chatId  int64
}

func NewReportService(chatService telegram.ChatService, chatId int64, db *sqlx.DB) ReportService {
	rep := NewDbReportRepository(db)
	return ReportService{
		rep:     &rep,
		chatSrv: chatService,
		chatId:  chatId,
	}

}

func (rs *ReportService) SendUsersReport() {
	report, err := rs.UsersLastMessage()
	if err != nil {
		log.Println(err)
	}
	rs.SendReport(report)
}

func (rs *ReportService) SendDailyMessagesReport() {
	now := time.Now().UTC()
	after := now.Truncate(24 * time.Hour)
	report, err := rs.UserMessagesCount(after)
	if err != nil {
		log.Println(err)
	}
	rs.SendReport(report)
}

func (rs *ReportService) SendWeeklyMessagesReport() {
	now := time.Now().UTC()
	report, err := rs.WeeklyMessagesCount(now)
	if err != nil {
		log.Println(err)
	}
	rs.SendReport(report)
}

func (rs *ReportService) SendDbSizeReport() {
	size, err := util.FileSizeInKib(storage.DbPath)
	if err != nil {
		log.Println("failed to get db size")
		return
	}
	var tmplBuf bytes.Buffer
	tmpl, err := template.New("db_size_report").Funcs(util.TemplateFuncMap).Parse(DbStats)
	if err != nil {
		log.Println(err)
		return
	}
	if err := tmpl.Execute(&tmplBuf, map[string]interface{}{
		"ReportDate": time.Now().UTC().Unix(),
		"DbSize":     size,
		"Threshold":  DbAlertThreshold,
	},
	); err != nil {
		log.Println(err)
		return
	}
	rs.SendReport(tmplBuf.String())

}

func (rs *ReportService) UsersLastMessage() (string, error) {
	userLastMessages, err := rs.rep.GetUserLastMessage()
	if err != nil {
		return "", err
	}
	var tmplBuf bytes.Buffer
	tmpl := template.Must(template.New("user_last_message_report").Funcs(util.TemplateFuncMap).Parse(UserLastMessageReport))
	if err := tmpl.Execute(&tmplBuf, userLastMessages); err != nil {
		return "", err
	}
	return tmplBuf.String(), nil
}
func (rs *ReportService) UserMessagesCount(after time.Time) (string, error) {
	userMessages, err := rs.rep.GetUserMessagesStat(after)
	if err != nil {
		return "", err
	}
	var tmplBuf bytes.Buffer
	tmpl := template.Must(template.New("user_messages_report_after").Funcs(util.TemplateFuncMap).Parse(UserMessagesAfterReport))
	if err := tmpl.Execute(&tmplBuf, struct {
		ReportDate int64
		Stat       []*UserMessageStat
	}{
		ReportDate: after.Unix(),
		Stat:       userMessages,
	},
	); err != nil {
		return "", err

	}
	return tmplBuf.String(), nil
}

func (rs *ReportService) WeeklyMessagesCount(reportDate time.Time) (string, error) {
	weekday := reportDate.Weekday()
	if weekday == 0 {
		weekday = 7
	}
	after := reportDate.AddDate(0, 0, -int(weekday)+1).Truncate(24 * time.Hour)
	return rs.UserMessagesCount(after)
}

func (rs *ReportService) SendReport(report string) {
	_, err := rs.chatSrv.SendMessage(telegram.MessageOut{Text: tgbotapi.EscapeText("Markdown", report), RepyToId: 0, ChatId: rs.chatId})
	if err != nil {
		log.Println("failed to send report via telegram")
	}

}
