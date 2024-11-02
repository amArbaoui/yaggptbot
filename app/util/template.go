package util

import (
	"html/template"
	"time"
)

var TemplateFuncMap = template.FuncMap{
	"formatDt": FormatReportDt,
}

func FormatReportDt(t int64) string {
	loc, _ := time.LoadLocation("Europe/Paris")
	return time.Unix(t, 0).In(loc).Format("2006-01-02 15:04:05 CET")
}
