package main

import (
	"fmt"
	"time"
	nodeClient "rcf/rcf-node-client"
)

func main() {
  // opening connection(tcp client) to node with id(port) 30
  client := nodeClient.NodeOpenConn(47)

  for {
	result := nodeClient.TopicPullStringData(client, 5, "altsensstring")

	println("string single results: ")
	fmt.Println(result)
	time.Sleep(1000*time.Microsecond)
  }

  // closing node conn at program end
  nodeClient.NodeCloseConn(client)
}
