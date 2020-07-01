package main

import (
	"fmt"
	"time"
	rcfNodeClient "rcf/rcfNodeClient"
)

func main() {
	// opening connection(tcp client) to node with id(port) 30
	client := rcfNodeClient.NodeOpenConn(47)

	for {
		result := rcfNodeClient.TopicPullStringData(client, 5, "altsensstring")

		println("single string results: ")
		fmt.Println(result)
		time.Sleep(10 * time.Millisecond)
	}

	// closing node conn at program end
	rcfNodeClient.NodeCloseConn(client)
}
