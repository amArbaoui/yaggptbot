package report

import (
	"amArbaoui/yaggptbot/app/telegram"
	"amArbaoui/yaggptbot/app/util"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestWeeklyReportDate(t *testing.T) {
	var tests = []struct {
		startTime    time.Time
		expectedDate time.Time
	}{
		{time.Date(2024, 11, 5, 19, 0, 0, 0, time.UTC),
			time.Date(2024, 11, 4, 0, 0, 0, 0, time.UTC),
		},
		{time.Date(2024, 11, 3, 19, 0, 0, 0, time.UTC),
			time.Date(2024, 10, 28, 0, 0, 0, 0, time.UTC),
		},
	}

	rs := ReportService{TgReportRepositoryStub{}, &telegram.ChatServiceStub{}, 12345}

	for _, tt := range tests {
		testname := fmt.Sprintf("starting %s report date should be %s", tt.startTime, tt.expectedDate)
		expectedDt := util.FormatReportDt(tt.expectedDate.Unix())
		t.Run(testname, func(t *testing.T) {
			weeklyReport, err := rs.WeeklyMessagesCount(tt.startTime)
			if err != nil || !strings.Contains(weeklyReport, expectedDt) {
				t.Errorf("string should contain %s.Report is %s", expectedDt, weeklyReport)
			}

		})
	}
}
