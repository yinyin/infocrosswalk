package slackwebhook

import "bytes"
import "strconv"
import "strings"
import "net/http"
import "net/url"
import "encoding/json"

import "github.com/yinyin/infocrosswalk"
import "github.com/yinyin/infocrosswalk/outgoing"

type slackAdapter struct {
	httpClient    *http.Client
	httpTransport *http.Transport
	botUserName   string
	hookUrl       string
	bufferSize    int
	messageBuffer []string
	overflowedBuffer bool
}

func getProxyFunction(proxyUrl string) (f func(*http.Request) (*url.URL, error), err error) {
	if "" == proxyUrl {
		return nil, nil
	}
	fixProxyUrl, err := url.Parse(proxyUrl)
	if nil != err {
		return nil, err
	}
	f = http.ProxyURL(fixProxyUrl)
	return f, nil
}

func NewAdapter(botUserName string, hookUrl string, proxyUrl string, bufferSize int) (adapter outgoing.Adapter, err error) {
	proxyFunc, err := getProxyFunction(proxyUrl)
	if nil != err {
		return nil, err
	}
	t := &http.Transport{Proxy: proxyFunc}
	c := &http.Client{Transport: t}
	var b []string
	if 0 < bufferSize {
		b = make([]string, 0, bufferSize)
	}
	adapter = &slackAdapter{c, t, botUserName, hookUrl, bufferSize, b, false}
	return adapter, nil
}

type slackMessageAttachment struct {
	FallbackText string `json:"fallback"`
	Color string `json:"color,omitempty"`
	PreText string `json"pretext,omitempty"`
	Text string `json:"text,omitempty"`
}

type slackMessagePayload struct {
	UserName  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	Attachments []slackMessageAttachment `json:"attachments,omitempty"`
}

func escapeText(v string) (result string) {
	result = strings.Replace(v, "&", "&amp;", -1)
	result = strings.Replace(result, "<", "&lt;", -1)
	result = strings.Replace(result, ">", "&gt;", -1)
	return result
}

func buildMessageText(channel string, tag string, text string, linkurl string) (message string) {
	message = escapeText(text)
	channel = escapeText(channel)
	tag = escapeText(tag)
	linkurl = escapeText(linkurl)
	if "" == tag {
		if "" != linkurl {
			message = "<" + linkurl + "|" + message + ">"
		}
	} else {
		if "" != linkurl {
			message = "<" + linkurl + "|" + tag + ">: " + message
		} else {
			message = tag + ": " + message
		}
	}
	if "" != channel {
		message = "[" + channel + "] " + message
	}
	if "" == message {
		message = "??? (empty message, something wrong...)"
	}
	return message
}

func (c *slackAdapter) sendContent(message []string) (err error) {
	l := len(message)
	var failText string
	var preText string
	var contentText string
	if l > 1 {
		failText = message[0] + " ... (" + strconv.Itoa(l) + " messages)"
		preText = "Have " + strconv.Itoa(l) + " messages ..."
		contentText = strings.Join(message, "\n")
	} else {
		failText = message[0]
		contentText = message[0]
	}
	p := slackMessagePayload{
		UserName:  c.botUserName,
		IconEmoji: "",
		Attachments: []slackMessageAttachment{slackMessageAttachment{
			FallbackText: failText,
			PreText: preText,
			Text: contentText,}},}
	b, err := json.Marshal(p)
	postbody := bytes.NewReader(b)
	resp, err := c.httpClient.Post(c.hookUrl, "application/json", postbody)
	defer resp.Body.Close()
	return err
}

func (c *slackAdapter) AddMessage(content *infocrosswalk.MessageContent) (err error) {
	message := buildMessageText(content.Channel, content.Tag, content.Text, content.ResourceUrl)
	if c.bufferSize > 0 {
		if c.bufferSize == len(c.messageBuffer) {
			c.overflowedBuffer = true
		} else {
			c.messageBuffer = append(c.messageBuffer, message)
		}
		return nil
	} else {
		return c.sendContent([]string{message});
	}
}

func (c *slackAdapter) Flush() (err error) {
	if len(c.messageBuffer) > 0 {
		err = c.sendContent(c.messageBuffer)
		c.messageBuffer = make([]string, 0, c.bufferSize)
	} else {
		err = nil
	}
	return err
}

func (c *slackAdapter) Close() {
	c.httpTransport.CloseIdleConnections()
}
