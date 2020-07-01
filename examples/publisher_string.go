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
	rcfNodeClient.TopicCreate(client, "altsensstring")

	// loop to create sample data which is pushed to topic
	for {
		// generating random int
		alt := rand.Intn(200)
		
		// pushing alt value to node, encoded as string. every sent string/ alt value represents one element/ msg in the topic
		rcfNodeClient.TopicPublishStringData(client, "altsensstring", strconv.Itoa(alt))

		// equals 10 & 100 Hz
		// time.Sleep(100 * time.Millisecond)
		time.Sleep(10 * time.Millisecond)
	}

	// closing node conn at program end
	rcfNodeClient.NodeCloseConn(client)
}
