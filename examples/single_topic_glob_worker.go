package main

import (
	"fmt"
	"time"
	nodeClient "rcf/rcf-node-client"
)

func main() {
  // opening connection(tcp client) to node with id(port) 30
  connChan, conn := nodeClient.NodeOpenConn(47)

  for {
	result := nodeClient.TopicPullGlobData(conn, connChan, 5, "altsensglob")

	println("glob single pull results: ")
	fmt.Println(result)
	time.Sleep(10000*time.Microsecond)
  }

  // closing node conn at program end
  nodeClient.NodeCloseConn(conn)
}
