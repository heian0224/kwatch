package wechat

import (
	"bytes"
	"fmt"
	"github.com/abahmed/kwatch/constant"
	"github.com/abahmed/kwatch/event"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
)

type Wechat struct {
	webhook string
	title   string
}

type wechatWebhookContent struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// NewWechat returns new Wechat web bot instance
func NewWechat(config map[string]string) *Wechat {
	webhook, ok := config["webhook"]
	if !ok || len(webhook) == 0 {
		logrus.Warnf("initializing Wechat with empty webhook url")
		return nil
	}

	logrus.Infof("initializing Wechat with webhook url: %s", webhook)

	return &Wechat{
		webhook: webhook,
		title:   config["title"],
	}
}

// Name returns name of the provider
func (r *Wechat) Name() string {
	return "Wechat"
}

// SendEvent sends event to the provider
func (r *Wechat) SendEvent(e *event.Event) error {
	return r.sendByWechatApi(r.buildRequestBodyWechat(e, ""))
}

func (r *Wechat) sendByWechatApi(reqBody string) error {
	client := &http.Client{}
	buffer := bytes.NewBuffer([]byte(reqBody))
	request, err := http.NewRequest(http.MethodPost, r.webhook, buffer)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	//reqDump, err := httputil.DumpRequestOut(request, true)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("REQUEST:\n%s", string(reqDump))
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	//fmt.Printf("response body:%s", body)
	if response.StatusCode != 200 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf(
			"call to Wechat alert returned status code %d: %s",
			response.StatusCode,
			string(body))
	}

	return nil
}

// SendMessage sends text message to the provider
func (r *Wechat) SendMessage(msg string) error {
	return r.sendByWechatApi(
		r.buildRequestBodyWechat(new(event.Event), msg),
	)
}

func (r *Wechat) buildRequestBodyWechat(
	e *event.Event,
	customMsg string) string {
	// add events part if it exists
	eventsText := constant.DefaultEvents
	events := strings.TrimSpace(e.Events)
	if len(events) > 0 {
		eventsText = e.Events
	}

	// add logs part if it exist
	logsText := constant.DefaultLogs
	logs := strings.TrimSpace(e.Logs)
	if len(logs) > 0 {
		logsText = e.Logs
	}

	// build text will be sent in the message use custom text if it's provided,
	// otherwise use default
	text := ""
	if len(customMsg) <= 0 {
		text = fmt.Sprintf(
			"**Pod:** %s\n"+
				"**Container:** %s\n"+
				"**Namespace:** %s\n"+
				"**Reason:** %s\n"+
				"**Events:**\n```\n%s\n```\n"+
				"**Logs:**\n```\n%s\n```",
			e.Name,
			e.Container,
			e.Namespace,
			e.Reason,
			eventsText,
			logsText,
		)
	} else {
		text = customMsg
	}
	var content = "# " + r.title + "\n" + text

	body := "{ \"msgtype\": \"markdown\", \"markdown\": { \"content\": \"" + content + "\" }}"
	return body
}
