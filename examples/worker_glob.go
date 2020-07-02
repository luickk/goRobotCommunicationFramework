package main

import (
	"fmt"
    "os"
	"math/rand"
	rcfNodeClient "rcf/rcfNodeClient"
	"strconv"
	"time"
)

func main() {
	// opening connection(tcp client) to node with id(port) 30
	client := rcfNodeClient.NodeOpenConn(47)

	// initiating topic listener
	// returns channel which every new incoming element/ msg is pushed to
	topicListener := rcfNodeClient.TopicGlobDataSubscribe(client, os.Args[1])
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
		case globData := <-topicListener:
			// converting altitude element/ msg which is encoded as string to integer
			// removing spaces before
			alti, _ := strconv.Atoi(globData["alt"])
			// printing new altitude, pushed to topic
			fmt.Println("sub Altitude glob changed: ", alti)

			// checking if new altitude is equal tp 99 for example purposes
			receivedMsgs++
			if alti == 99 {
				rand := strconv.Itoa(rand.Intn(1000000))
				receivedMsgs++
				go func() {
					res := rcfNodeClient.ServiceExec(client, "testServiceDelay", []byte("randTestParamFromGlobWorker"+rand))
					println("results delay: " + string(res))
				}()
			}
		}
	}

	// closing node conn at program end
	rcfNodeClient.NodeCloseConn(client)
}
