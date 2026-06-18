package domain

type ReminderJobResult struct {
	UsersProcessed           int `json:"users_processed"`
	DevicesNotified          int `json:"devices_notified"`
	NotificationsSent        int `json:"notifications_sent"`
	SendFailures              int `json:"send_failures"`
	StaleTokensRemoved        int `json:"stale_tokens_removed"`
	UsersSkipped              int `json:"users_skipped"`
	DevicesSkippedOutsideHour int `json:"devices_skipped_outside_hour"`
}
