package incoming

import "github.com/yinyin/infocrosswalk"

type Adapter interface {
	FetchMessage(out chan<- infocrosswalk.MessageContent) (err error)
	Close()
}
