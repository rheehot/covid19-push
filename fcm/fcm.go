package fcm

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"golang.org/x/net/context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"

	"github.com/sukso96100/covid19-push/database"
)

type FCMObject struct {
	App       *firebase.App
	MsgClient *messaging.Client
	Ctx       context.Context
}

var fcmApp *FCMObject

func InitFCMApp(credential string) {
	fcmApp = &FCMObject{}
	fcmApp.Init(credential)
}

func GetFCMApp() *FCMObject {
	return fcmApp
}

func (fcm *FCMObject) Init(credential string) {
	fcm.Ctx = context.Background()
	app, err := firebase.NewApp(fcm.Ctx, nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	fcm.App = app

	client, err := fcm.App.Messaging(fcm.Ctx)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	fcm.MsgClient = client
}

func (fcm *FCMObject) PushStatData(current database.StatData) {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: "코로나19 발생 현황",
			Body:  database.CreateStatMsg(current),
		},
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				RequireInteraction: true,
			},
			FcmOptions: &messaging.WebpushFcmOptions{
				Link: createNotificationUrl("http://ncov.mohw.go.kr/index.jsp"),
			},
		},
		Topic: "stat",
	}

	// Send a message to the devices subscribed to the provided topic.
	response, err := fcm.MsgClient.Send(fcm.Ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Successfully sent stat message:", response)

}

func (fcm *FCMObject) PushNewsData(newsData database.NewsData) {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: newsData.Title,
			Body:  newsData.Department,
		},
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				RequireInteraction: true,
			},
			FcmOptions: &messaging.WebpushFcmOptions{
				Link: createNotificationUrl(newsData.Link),
			},
		},
		Topic: "news",
	}
	// Send a message to the devices subscribed to the provided topic.
	response, err := fcm.MsgClient.Send(fcm.Ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Successfully sent news message:", response)
}

func (fcm *FCMObject) SendConfirmNotification(token string, isSubscribe bool, topic string) {
	var title string
	var body string
	var topicDisplay string
	if topic == "stat" {
		topicDisplay = "발생 현황"
	} else if topic == "news" {
		topicDisplay = "공지사항"
	}
	if isSubscribe {
		title = "코로나19 알리미 구독 완료"
		body = fmt.Sprintf(
			"질병관리본부 코로나19 홈페이지의 %s을 푸시알림으로 알려드립니다.",
			topicDisplay,
		)
	} else {
		title = fmt.Sprintf("코로나19 알리미 %s 구독 해제됨", topicDisplay)
		body = "알림을 수신하지 않으려면 웹 브라우저에서 알림 권한을 차단하세요."
	}
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: token,
	}

	// Send a message to the devices subscribed to the provided topic.
	response, err := fcm.MsgClient.Send(fcm.Ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Successfully sent confirm message:", response)
}

func createNotificationUrl(url string) string {
	hostname := os.Getenv("APP_HOST")
	return fmt.Sprintf("https://%s/redirect/%s",
		hostname, base64.URLEncoding.EncodeToString([]byte(url)))
}
