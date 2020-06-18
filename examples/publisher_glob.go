package main

import (
  // "time"
	"math/rand"
	"strconv"
	rcfNodeClient "rcf/rcfNodeClient"
)

func main() {
  // opening connection(tcp client) to node with id(port) 30
  client := rcfNodeClient.NodeOpenConn(47)

  // creating topic by sending cmd to node
  rcfNodeClient.TopicCreate(client, "altsensglob")

  // loop to create sample data which is pushed to topic
  for {
    // generating random int
    alt := rand.Intn(200)
    // printing sample data
    // fmt.Println(alt)
    // putting sample data into map
    dataMap := make(map[string]string)
    dataMap["alt"] = strconv.Itoa(alt)
    // pushing alt value to node, encoded as string. every sent string/ alt value represents one element/ msg in the topic
    rcfNodeClient.TopicPublishGlobData(client, "altsensglob", dataMap)
    // time.Sleep(1000*time.Microseccond)
    // time.Sleep(1*time.Second)
  }

  // closing node conn at program end
  rcfNodeClient.NodeCloseConn(client)
}
