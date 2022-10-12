package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/examples/newclient"
	"github.com/gammazero/nexus/v3/wamp"
)

const procedureName = "complex-call"

func main() {
	logger := log.New(os.Stdout, "CALLEE> ", 0)
	// Connect callee client with requested socket type and serialization.
	callee, err := newclient.NewClient(logger)
	if err != nil {
		logger.Fatal(err)
	}
	defer callee.Close()

	// Register procedure "sum"
	if err = callee.Register(procedureName, complexCall, nil); err != nil {
		logger.Fatal("Failed to register procedure:", err)
	}
	logger.Println("Registered procedure", procedureName, "with router")

	// Wait for CTRL-c or client close while handling remote procedure calls.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	select {
	case <-sigChan:
	case <-callee.Done():
		logger.Print("Router gone, exiting")
		return // router gone, just exit
	}

	if err = callee.Unregister(procedureName); err != nil {
		logger.Println("Failed to unregister procedure:", err)
	}
}

func complexCall(ctx context.Context, inv *wamp.Invocation) client.InvokeResult {
	var name string
	var ID int
	var logs []string
	var complex struct {
		Arg1  int    `wamp:"arg_1"`
		Value string `wamp:"value"`
	}
	if err := client.Unpack(inv.Arguments, &name, &ID, &logs, &complex); err != nil {
		log.Printf("cannot unpack arguments: %v", err)
		return client.ErrWithMessage(wamp.ErrInvalidArgument,
			fmt.Sprintf("incorrect arguments: %v", err))
	}
	return client.InvokeResult{Args: wamp.List{
		fmt.Sprintf("%v %v %v %+v",
			name, ID, logs, complex),
	}}
}
