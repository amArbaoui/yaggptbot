package report

import (
	"amArbaoui/yaggptbot/app/util"
)

func NewReportScheduler(repService *ReportService) (*util.Scheduler, error) {
	scheduler, err := util.NewScheduler()
	if err != nil {
		return nil, err
	}
	scheduler.AddCronJob(
		CronDailyReport,
		repService.SendUsersReport,
	)
	scheduler.AddCronJob(
		CronDailyReport,
		repService.SendDailyMessagesReport,
	)
	scheduler.AddCronJob(
		CronWeeklyReport,
		repService.SendWeeklyMessagesReport,
	)
	scheduler.AddCronJob(
		CronDailyReport,
		repService.SendDbSizeReport,
	)
	return scheduler, nil
}
