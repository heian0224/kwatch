package wechat

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abahmed/kwatch/event"
	"github.com/stretchr/testify/assert"
)

func TestEmptyConfig(t *testing.T) {
	assertions := assert.New(t)

	c := NewWechat(map[string]string{})
	assertions.Nil(c)
}

func TestRocketChat(t *testing.T) {
	assertions := assert.New(t)

	config := map[string]string{
		"webhook": "testtest",
	}
	c := NewWechat(config)
	assertions.NotNil(c)

	assertions.Equal(c.Name(), "Wechat")
}

func TestBuildRequestBodyWechat(t *testing.T) {
	assertions := assert.New(t)
	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"isOk": true}`))
		}))

	defer s.Close()

	config := map[string]string{
		"webhook": s.URL,
	}
	c := NewWechat(config)
	assertions.NotNil(c)
	ev := event.Event{
		Name:      "test-pod",
		Container: "test-container",
		Namespace: "default",
		Reason:    "OOMKILLED",
		Logs:      "test\ntestlogs",
		Events:    "test",
	}
	expectMessage := "{ \"msgtype\": \"markdown\", \"markdown\": { \"content\": \"# \n**Pod:** test-pod\n**Container:** test-container\n**Namespace:** default\n**Reason:** OOMKILLED\n**Events:**\n```\ntest\n```\n**Logs:**\n```\ntest\ntestlogs\n```\" }}"
	assertions.Equal(expectMessage, c.buildRequestBodyWechat(&ev, ""))
}

func TestSendMessage(t *testing.T) {
	assertions := assert.New(t)

	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"isOk": true}`))
		}))

	defer s.Close()

	config := map[string]string{
		"webhook": s.URL,
	}

	c := NewWechat(config)
	assertions.NotNil(c)

	assertions.Nil(c.SendMessage("test"))
}

func TestSendMessageError(t *testing.T) {
	assertions := assert.New(t)

	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))

	defer s.Close()

	config := map[string]string{
		"webhook": s.URL,
	}
	c := NewWechat(config)
	assertions.NotNil(c)

	assertions.NotNil(c.SendMessage("test"))
}

func TestSendEvent(t *testing.T) {
	assertions := assert.New(t)

	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"isOk": true}`))
		}))

	defer s.Close()

	config := map[string]string{
		"webhook": s.URL,
	}
	c := NewWechat(config)
	assertions.NotNil(c)

	ev := event.Event{
		Name:      "test-pod",
		Container: "test-container",
		Namespace: "default",
		Reason:    "OOMKILLED",
		Logs:      "test\ntestlogs",
		Events: "event1-event2-event3-event1-event2-event3-event1-event2-" +
			"event3\nevent5\nevent6-event8-event11-event12",
	}
	assertions.Nil(c.SendEvent(&ev))
}

func TestInvalidHttpRequest(t *testing.T) {
	assertions := assert.New(t)

	config := map[string]string{
		"webhook": "h ttp://localhost",
	}
	c := NewWechat(config)
	assertions.NotNil(c)

	assertions.NotNil(c.SendMessage("test"))

	config = map[string]string{
		"webhook": "http://localhost:132323",
	}
	c = NewWechat(config)
	assertions.NotNil(c)

	assertions.NotNil(c.SendMessage("test"))
}
