package kallithea

import "time"
import "strings"
import "regexp"
import "net/http"
import "crypto/tls"

import "github.com/yinyin/infocrosswalk"
import "github.com/yinyin/infocrosswalk/incoming"

type kallitheaAdapter struct {
	httpClient    *http.Client
	httpTransport *http.Transport
	atomUrl       string
}

func NewAdapter(atomUrl string) (adapter incoming.Adapter, err error) {
	t := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	c := &http.Client{Transport: t}
	adapter = &kallitheaAdapter{c, t, atomUrl}
	return adapter, nil
}

type linkUrlHref struct {
	Href string `xml:"href,attr"`
}

type changesetEntry struct {
	Title      string      `xml:"title"`
	LinkUrl    linkUrlHref `xml:"link"`
	UpdateTime string      `xml:"updated"`
	Summary    string      `xml:"summary"`
}

var (
	regexAuthor   = regexp.MustCompile("([A-Za-z0-9\\s,-_]+)(\\s*<[A-Za-z0-9\\s,-_<>@\\.]*)?\\s*committed on [0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}\\s*<")
	regexBranch   = regexp.MustCompile(">branch:\\s*([^<\\s]+)<")
	regexRevision = regexp.MustCompile("changeset: <[^>]+>([a-fA-F0-9]+)</a>")
)

func (e *changesetEntry) GetMessageContent() (channel string, tag string, message string, link string) {
	var author string
	m := regexAuthor.FindStringSubmatch(e.Summary)
	if nil != m {
		author = m[1]
	}
	m = regexBranch.FindStringSubmatch(e.Summary)
	if nil != m {
		channel = m[1]
	}
	m = regexRevision.FindStringSubmatch(e.Summary)
	if nil != m {
		tag = m[1]
	}
	m_aux := strings.SplitN(e.Title, "\n", 2)
	message = strings.Trim(m_aux[0], " \r\n\t")
	if "" != author {
		message = message + " - " + author
	}
	return channel, tag, message, e.LinkUrl.Href
}

func (e *changesetEntry) GetTime() (t time.Time) {
	t, err := time.Parse(time.RFC3339, e.UpdateTime)
	if nil != err {
		return time.Unix(0, 0)
	}
	return t
}

func (e *changesetEntry) Reset() {
	e.Title = ""
	e.LinkUrl.Href = ""
	e.UpdateTime = ""
	e.Summary = ""
}

func (c *kallitheaAdapter) FetchMessage(lastProgress time.Time, out chan<- infocrosswalk.MessageContent) (progress time.Time, err error) {
	resp, err := c.httpClient.Get(c.atomUrl)
	if nil != err {
		return lastProgress, err
	}
	defer resp.Body.Close()
	var e changesetEntry
	return incoming.DecodeAtom(lastProgress, out, resp.Body, &e)
}

func (c *kallitheaAdapter) Close() {
	c.httpTransport.CloseIdleConnections()
}
