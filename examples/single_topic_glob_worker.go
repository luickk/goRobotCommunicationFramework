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
		result := rcfNodeClient.TopicPullGlobData(client, 5, "altsensglob")

		println("single glob pull results: ")
		fmt.Println(result)
		time.Sleep(10 * time.Millisecond)
	}

	// closing node conn at program end
	rcfNodeClient.NodeCloseConn(client)
}
