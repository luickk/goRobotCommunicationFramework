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

	println("glob single pull results: ")
	fmt.Println(result)
	time.Sleep(10000*time.Microsecond)
    // time.Sleep(1*time.Second)
  }

  // closing node conn at program end
  rcfNodeClient.NodeCloseConn(client)
}
