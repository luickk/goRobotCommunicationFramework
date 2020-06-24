package main

import (
	"math/rand"
	rcfNodeClient "rcf/rcfNodeClient"
	"strconv"
	"time"
)

func main() {
	// opening connection(tcp client) to node with id(port) 30
	client := rcfNodeClient.NodeOpenConn(47)

	// creating topic by sending cmd to node
	rcfNodeClient.TopicCreate(client, "altsensglob")

	// loop to create sample data which is pushed to topic
	for {
		// generating random int
		alt := rand.Intn(200)
		// printing sample data
		// fmt.Println(alt)
		// putting sample data into map
		dataMap := make(map[string]string)
		dataMap["alt"] = strconv.Itoa(alt)
		// pushing alt value to node, encoded as string. every sent string/ alt value represents one element/ msg in the topic
		rcfNodeClient.TopicPublishGlobData(client, "altsensglob", dataMap)

		// euals 10 & 100 Hz
		// time.Sleep(100 * time.Millisecond)
		time.Sleep(10 * time.Millisecond)
	}

	// closing node conn at program end
	rcfNodeClient.NodeCloseConn(client)
}
