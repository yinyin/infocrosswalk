package outgoing

import "github.com/yinyin/infocrosswalk"

type Adapter interface {
	AddMessage(content *infocrosswalk.MessageContent) (err error)
	Flush() (err error)
	Close()
}
