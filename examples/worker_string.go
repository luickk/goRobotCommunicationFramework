package main

import (
	"fmt"
	rcfNodeClient "rcf/rcfNodeClient"
	"strconv"
	"strings"
	"time"
)

func main() {
	// opening connection(tcp client) to node with id(port) 30
	client := rcfNodeClient.NodeOpenConn(47)

	// initiating topic listener
	// returns channel which every new incoming element/ msg is pushed to
	topicListener := rcfNodeClient.TopicStringDataSubscribe(client, "altsensstring")
	receivedMsgs := 0

	go func() {
		for {
			time.Sleep(1 * time.Second)

			println("Benchmark, received msgs in one second: " + strconv.Itoa(receivedMsgs))

			receivedMsgs = 0
		}
	}()
	// smaple loop
	for {
		// select statement to wait for new incoming elements/msgs from listened to topic
		select {
		// if new element/ msg was pushed to listened topic, it is also pushed to the listener channel
		case alt := <-topicListener:
			// converting altitude element/ msg which is encoded as string to integer
			// removing spaces before
			alti, _ := strconv.Atoi(strings.TrimSpace(alt))
			// printing new altitude, pushed to topic
			fmt.Println("Altitude string changed: ", alti)
			receivedMsgs++
		}
	}

	// closing node conn at program end
	rcfNodeClient.NodeCloseConn(client)
}
