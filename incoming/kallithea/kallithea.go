package kallithea

import "strings"
import "regexp"
import "net/http"
import "crypto/tls"
import "encoding/xml"

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
	Href    string `xml:"href,attr"`
}

type changesetEntry struct {
	Title      string `xml:"title"`
	LinkUrl    linkUrlHref `xml:"link"`
	UpdateTime string `xml:"updated"`
	Summary    string `xml:"summary"`
}

var (
	regexAuthor = regexp.MustCompile("([A-Za-z0-9\\s,-_]+)(\\s*<[A-Za-z0-9\\s,-_<>@\\.]*)?\\s*committed on [0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}\\s*<")
	regexBranch = regexp.MustCompile(">branch:\\s*([^<\\s]+)<")
	regexRevision = regexp.MustCompile("changeset: <[^>]+>([a-fA-F0-9]+)</a>")
)

func translateMessage(msg string, v string) (channel string, tag string, message string) {
	var author string
	m := regexAuthor.FindStringSubmatch(v)
	if nil != m {
		author = m[1]
	}
	m = regexBranch.FindStringSubmatch(v)
	if nil != m {
		channel = m[1]
	}
	m = regexRevision.FindStringSubmatch(v)
	if nil != m {
		tag = m[1]
	}
	m_aux := strings.SplitN(msg, "\n", 2)
	message = strings.Trim(m_aux[0], " \r\n\t")
	if "" != author {
		message = message + " - " + author
	}
	return channel, tag, message
}

func (c *kallitheaAdapter) FetchMessage(out chan<- infocrosswalk.MessageContent) (err error) {
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
				var e changesetEntry
				decoder.DecodeElement(&e, &se)
				channel, tag, message := translateMessage(e.Title, e.Summary)
				m := infocrosswalk.MessageContent{channel, tag, message, e.LinkUrl.Href}
				out <- m
			}
		}
	}
	return nil
}

func (c *kallitheaAdapter) Close() {
}
