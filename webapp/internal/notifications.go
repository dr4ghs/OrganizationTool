package internal

import (
	"fmt"
	"net/http"
)

type NotificationType int

const (
	InfoNotification NotificationType = iota
	WarningNotification
	ErrorNotification
)

func AddNotification(w http.ResponseWriter, typ NotificationType, message string) {
	w.Header().Add("X-Notification", fmt.Sprintf("%d", typ))
	w.Header().Add("X-Notification-Message", message)
}
