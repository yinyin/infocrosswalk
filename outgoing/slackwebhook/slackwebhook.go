package slackwebhook

import "bytes"
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

func NewAdapter(botUserName string, hookUrl string, proxyUrl string) (adapter outgoing.Adapter, err error) {
	proxyFunc, err := getProxyFunction(proxyUrl)
	if nil != err {
		return nil, err
	}
	t := &http.Transport{Proxy: proxyFunc}
	c := &http.Client{Transport: t}
	adapter = &slackAdapter{c, t, botUserName, hookUrl}
	return adapter, nil
}

type slackMessagePayload struct {
	UserName  string `json:"username,omitempty"`
	Text      string `json:"text"`
	IconEmoji string `json:"icon_emoji,omitempty"`
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

func (c *slackAdapter) AddMessage(content *infocrosswalk.MessageContent) (err error) {
	message := buildMessageText(content.Channel, content.Tag, content.Text, content.ResourceUrl)
	p := slackMessagePayload{
		UserName:  c.botUserName,
		Text:      message,
		IconEmoji: ""}
	b, err := json.Marshal(p)
	postbody := bytes.NewReader(b)
	resp, err := c.httpClient.Post(c.hookUrl, "application/json", postbody)
	defer resp.Body.Close()
	return err
}

func (c *slackAdapter) Close() {
	c.httpTransport.CloseIdleConnections()
}
