package flow

import "fmt"

import "github.com/yinyin/infocrosswalk"
import "github.com/yinyin/infocrosswalk/outgoing"
import "github.com/yinyin/infocrosswalk/incoming"

func Run(outAdapter outgoing.Adapter, inAdapter incoming.Adapter, bufferSize int) (errorCount int) {
	if bufferSize < 1 {
		bufferSize = 3
	}

	flowPipe := make(chan infocrosswalk.MessageContent, bufferSize)

	go func() {
		err := inAdapter.FetchMessage(flowPipe)
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
	return errorCount
}
