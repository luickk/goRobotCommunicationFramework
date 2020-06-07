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
	result := nodeClient.TopicPullGlobData(client, 5, "altsensglob")

	println("glob single pull results: ")
	fmt.Println(result)
	time.Sleep(10000*time.Microsecond)
    // time.Sleep(1*time.Second)
  }

  // closing node conn at program end
  nodeClient.NodeCloseConn(client)
}
