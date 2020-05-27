package main

import (
  "time"
	"math/rand"
	"strconv"
	nodeClient "rcf/rcf-node-client"
)

func main() {
  // opening connection(tcp client) to node with id(port) 30
  _, conn := nodeClient.NodeOpenConn(47)

  // creating topic by sending cmd to node
  nodeClient.TopicCreate(conn, "altsensstring")

  // loop to create sample data which is pushed to topic
  for {
    // generating random int
    alt := rand.Intn(100)
    // printing sample data
    // fmt.Println(alt)
    // pushing alt value to node, encoded as string. every sent string/ alt value represents one element/ msg in the topic
    nodeClient.TopicPublishStringData(conn, "altsensstring", strconv.Itoa(alt))
    time.Sleep(1*time.Second)
  }

  // closing node conn at program end
  nodeClient.NodeCloseConn(conn)
}
