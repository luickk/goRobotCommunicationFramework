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
	result := nodeClient.TopicPullStringData(conn, connChan, 5, "altsensstring")

	println("string single results: ")
	fmt.Println(result)
	time.Sleep(1000*time.Microsecond)
  }

  // closing node conn at program end
  nodeClient.NodeCloseConn(conn)
}
