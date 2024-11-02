package report

import "time"

type TgReportRepositoryStub struct {
}

func (tgrep TgReportRepositoryStub) GetUserLastMessage() ([]*UserLastMessage, error) {
	return []*UserLastMessage{}, nil
}

func (tgrep TgReportRepositoryStub) GetUserMessagesStat(after time.Time) ([]*UserMessageStat, error) {
	return []*UserMessageStat{}, nil
}
