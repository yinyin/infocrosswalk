package flow

import "fmt"
import "time"

import "github.com/yinyin/infocrosswalk"
import "github.com/yinyin/infocrosswalk/outgoing"
import "github.com/yinyin/infocrosswalk/incoming"

func Run(lastProgress time.Time, outAdapter outgoing.Adapter, inAdapter incoming.Adapter, bufferSize int) (resultProgress time.Time, errorCount int) {
	if bufferSize < 1 {
		bufferSize = 3
	}

	flowPipe := make(chan infocrosswalk.MessageContent, bufferSize)
	resultProgress = lastProgress

	go func() {
		var err error
		resultProgress, err = inAdapter.FetchMessage(lastProgress, flowPipe)
		if nil != err {
			fmt.Println("error (in: ", inAdapter, "):", err)
		}
		close(flowPipe)
	}()

	errorCount = 0
	for m := range flowPipe {
		err := outAdapter.AddMessage(&m)
		if nil != err {
			fmt.Println("error (out: ", outAdapter, "):", err)
			errorCount = errorCount + 1
		}
	}
	return resultProgress, errorCount
}
