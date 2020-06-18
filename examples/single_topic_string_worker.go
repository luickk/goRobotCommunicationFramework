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

	println("string single results: ")
	fmt.Println(result)
	time.Sleep(1000*time.Microsecond)
    // time.Sleep(1*time.Second)
  }

  // closing node conn at program end
  rcfNodeClient.NodeCloseConn(client)
}
