package incoming

import "time"

import "github.com/yinyin/infocrosswalk"

type Adapter interface {
	FetchMessage(lastProgress time.Time, out chan<- infocrosswalk.MessageContent) (progress time.Time, err error)
	Close()
}
