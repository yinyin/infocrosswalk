package redmine

import "time"
import "net/http"
import "crypto/tls"

import "github.com/yinyin/infocrosswalk"
import "github.com/yinyin/infocrosswalk/incoming"

type redmineAdapter struct {
	httpClient    *http.Client
	httpTransport *http.Transport
	atomUrl       string
}

func NewAdapter(atomUrl string) (adapter incoming.Adapter, err error) {
	t := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	c := &http.Client{Transport: t}
	adapter = &redmineAdapter{c, t, atomUrl}
	return adapter, nil
}

type linkUrlHref struct {
	Href string `xml:"href,attr"`
}

type authorMeta struct {
	Name string `xml:"name"`
}

type activityEntry struct {
	Title      string      `xml:"title"`
	LinkUrl    linkUrlHref `xml:"link"`
	UpdateTime string      `xml:"updated"`
	Author     authorMeta  `xml:"author"`
}

func (e *activityEntry) GetMessageContent() (channel string, tag string, message string, link string) {
	channel = e.Author.Name
	message = e.Title
	link = e.LinkUrl.Href
	return channel, tag, message, link
}

func (e *activityEntry) GetTime() (t time.Time) {
	t, err := time.Parse(time.RFC3339, e.UpdateTime)
	if nil != err {
		return time.Unix(0, 0)
	}
	return t
}

func (e *activityEntry) Reset() {
	e.Title = ""
	e.LinkUrl.Href = ""
	e.UpdateTime = ""
	e.Author.Name = ""
}

func (c *redmineAdapter) FetchMessage(lastProgress time.Time, out chan<- infocrosswalk.MessageContent) (progress time.Time, err error) {
	resp, err := c.httpClient.Get(c.atomUrl)
	if nil != err {
		return lastProgress, err
	}
	defer resp.Body.Close()
	var e activityEntry
	return incoming.DecodeAtom(lastProgress, out, resp.Body, &e)
}

func (c *redmineAdapter) Close() {
	c.httpTransport.CloseIdleConnections()
}
