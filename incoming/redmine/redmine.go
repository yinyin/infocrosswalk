package redmine

import "net/http"
import "crypto/tls"
import "encoding/xml"

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
	Href    string `xml:"href,attr"`
}

type authorMeta struct {
	Name  string `xml:"name"`
}

type activityEntry struct {
	Title      string `xml:"title"`
	LinkUrl    linkUrlHref `xml:"link"`
	UpdateTime string `xml:"updated"`
	Author    authorMeta `xml:"author"`
}

func (e *activityEntry) getMessageContent() (channel string, tag string, message string, link string) {
	channel = e.Author.Name
	message = e.Title
	link = e.LinkUrl.Href
	return channel, tag, message, link
}

func (c *redmineAdapter) FetchMessage(out chan<- infocrosswalk.MessageContent) (err error) {
	resp, err := http.Get(c.atomUrl)
	if nil != err {
		return err
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	for {
		t, _ := decoder.Token()
		if nil == t {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "entry" {
				var e activityEntry
				decoder.DecodeElement(&e, &se)
				channel, tag, message, link := e.getMessageContent()
				m := infocrosswalk.MessageContent{channel, tag, message, link}
				out <- m
			}
		}
	}
	return nil
}

func (c *redmineAdapter) Close() {
}
