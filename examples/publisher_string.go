package main

import (
	"math/rand"
	rcfNodeClient "rcf/rcfNodeClient"
	"strconv"
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
		// printing sample data
		// fmt.Println(alt)
		// pushing alt value to node, encoded as string. every sent string/ alt value represents one element/ msg in the topic
		rcfNodeClient.TopicPublishStringData(client, "altsensstring", strconv.Itoa(alt))
		// time.Sleep(1000 * time.Microsecond)
		// time.Sleep(1*time.Second)
	}

	// closing node conn at program end
	rcfNodeClient.NodeCloseConn(client)
}
