package main

import (
  "time"
	"math/rand"
	"strconv"
	nodeClient "rcf/rcf-node-client"
)

func main() {
  // opening connection(tcp client) to node with id(port) 30
  client := nodeClient.NodeOpenConn(47)

  // creating topic by sending cmd to node
  nodeClient.TopicCreate(client, "altsensglob")

  // loop to create sample data which is pushed to topic
  for {
    // generating random int
    alt := rand.Intn(100)
    // printing sample data
    // fmt.Println(alt)
    // putting sample data into map
    dataMap := make(map[string]string)
    dataMap["alt"] = strconv.Itoa(alt)
    // pushing alt value to node, encoded as string. every sent string/ alt value represents one element/ msg in the topic
    nodeClient.TopicPublishGlobData(client, "altsensglob", dataMap)
    time.Sleep(1000*time.Microsecond)
  }

  // closing node conn at program end
  nodeClient.NodeCloseConn(client)
}
