package incoming

import "time"
import "io"
import "encoding/xml"

import "github.com/yinyin/infocrosswalk"

type Entry interface {
	GetMessageContent() (channel string, tag string, message string, link string)
	GetTime() (t time.Time)
	Reset()
}


func DecodeAtom(lastProgress time.Time, out chan<- infocrosswalk.MessageContent, r io.Reader, e Entry) (progress time.Time, err error) {
	decoder := xml.NewDecoder(r)
	progress = lastProgress
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
				tstamp := e.GetTime()
				if tstamp.After(lastProgress) {
					if tstamp.After(progress) {
						progress = tstamp
					}
					channel, tag, message, link := e.GetMessageContent()
					m := infocrosswalk.MessageContent{channel, tag, message, link}
					out <- m
				}
			}
		}
	}
	return progress, nil
}
