package helpers

import (
	"os"

	"github.com/appleboy/go-fcm"
	"github.com/sirupsen/logrus"
)

func SendNotification(deviceToken, title, message, deeplink string, path, data *string) error {
	var notification *fcm.Message

	if path != nil && data != nil {
		notification = &fcm.Message{
			To: deviceToken,
			Data: map[string]interface{}{
				"path":     &path,
				"data":     &data,
				"deeplink": deeplink,
			},
			Notification: &fcm.Notification{
				Title: title,
				Body:  message,
				Badge: "1",
			},
		}
	} else {
		notification = &fcm.Message{
			To: deviceToken,
			Data: map[string]interface{}{
				"deeplink": deeplink,
			},
			Notification: &fcm.Notification{
				Title: title,
				Body:  message,
				Badge: "1",
			},
		}
	}

	client, err := fcm.NewClient(os.Getenv("FCM_KEY"))
	if err != nil {
		logrus.Error(err.Error(), err)
	}

	response, err := client.Send(notification)
	if err != nil {
		logrus.Error(err.Error(), err)
	}

	return response.Error
}
