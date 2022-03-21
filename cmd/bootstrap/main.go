package main

import (
	"fmt"
	"github.com/iovisor/gobpf/bcc"
	"log"
)

func main() {
	hook := `
#include <uapi/linux/ptrace.h>
#include <linux/string.h>

BPF_PERF_OUTPUT(events);

inline int function_was_called(struct pt_regs *ctx) {

   char x[29] = "Hey, the handler was called!";
   events.perf_submit(ctx, &x, sizeof(x));
   return 0;
}`
	bpfModule := bcc.NewModule(hook, []string{})

	uprobeFd, err := bpfModule.LoadUprobe("function_was_called")
	if err != nil {
		log.Fatal(err)
	}

	err = bpfModule.AttachUprobe("/Users/mo/Documents/dev/salty/cmd/server/server", "main.handlerFunction", uprobeFd, -1)
	if err != nil {
		log.Fatal(err)
	}

	table := bcc.NewTable(bpfModule.TableId("events"), bpfModule)

	outputChannel := make(chan []byte)

	perfMap, err := bcc.InitPerfMap(table, outputChannel, nil)
	if err != nil {
		log.Fatal(err)
	}

	perfMap.Start()
	defer perfMap.Stop()

	for {
		value := <-outputChannel
		fmt.Println(string(value))
	}
}
