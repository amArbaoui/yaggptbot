package report

const DbAlertThreshold int64 = 1024 * 100 // 100 mib

const UserLastMessageReport = `
🙉 Active users
{{- range . }}
{{ .TgName }} (last message {{ .LastMessage | formatDt }})
{{- end }}
`

const UserMessagesAfterReport = `
✉️ Messages after {{ .ReportDate | formatDt }}
{{- range .Stat }}
{{ .TgName }} - {{ .MessageQt }}
{{- end }}
`

const DbStats = `
Bot DB as of {{ .ReportDate | formatDt }}
{{- if le .DbSize .Threshold }}
✅ {{ .DbSize }} Kib
{{- else }}
❌ {{ .DbSize }} Kib
{{- end }}
`

const CronDailyReport = "0 17 * * *"
const CronWeeklyReport = "0 17 * * SUN"
