package main

import (
	"time"
	"math/rand"
	"strconv"
	nodeClient "rcf/rcf-node-client"
)

func main() {
  // opening connection(tcp client) to node with id(port) 47
  client := nodeClient.NodeOpenConn(47)

  // creating topic by sending cmd to node
  nodeClient.TopicCreate(client, "altsensmglob")
  nodeClient.TopicCreate(client, "radarsensmglob")

  // loop to create sample data which is pushed to topic
  for {
    // generating random int
    random := rand.Intn(100)
    // printing sample data
    // fmt.Println(random)
    // putting sample data into map
    radDataMap := make(map[string]string)
    radDataMap["rad"] = strconv.Itoa(random)

    altDataMap := make(map[string]string)
    altDataMap["alt"] = strconv.Itoa(random)
    // pushing alt value to node, encoded as string. every sent string/ alt value represents one element/ msg in the topic
    nodeClient.TopicPublishGlobData(client, "radarsensmglob", radDataMap)
    nodeClient.TopicPublishGlobData(client, "altsensmglob", altDataMap)

    time.Sleep(1000*time.Microsecond)
    // time.Sleep(1*time.Second)
  }

  // closing node conn at program end
  nodeClient.NodeCloseConn(client)
}
