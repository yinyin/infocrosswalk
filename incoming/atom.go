package incoming

import "io"
import "encoding/xml"

import "github.com/yinyin/infocrosswalk"

type Entry interface {
	GetMessageContent() (channel string, tag string, message string, link string)
	Reset()
}

func DecodeAtom(out chan<- infocrosswalk.MessageContent, r io.Reader, e Entry) (err error) {
	decoder := xml.NewDecoder(r)
	for {
		t, _ := decoder.Token()
		if nil == t {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "entry" {
				e.Reset()
				decoder.DecodeElement(e, &se)
				channel, tag, message, link := e.GetMessageContent()
				m := infocrosswalk.MessageContent{channel, tag, message, link}
				out <- m
			}
		}
	}
	return nil
}
