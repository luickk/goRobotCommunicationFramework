package main

import (
	"time"
	"math/rand"
	"strconv"
	rcfNodeClient "rcf/rcfNodeClient"
)

func main() {
  // opening connection(tcp client) to node with id(port) 47
  client := rcfNodeClient.NodeOpenConn(47)

  // creating topic by sending cmd to node
  rcfNodeClient.TopicCreate(client, "altsensmglob")
  rcfNodeClient.TopicCreate(client, "radarsensmglob")

  // loop to create sample data which is pushed to topic
  for {
    // generating random int
    random := rand.Intn(200)
    // printing sample data
    // fmt.Println(random)
    // putting sample data into map
    radDataMap := make(map[string]string)
    radDataMap["rad"] = strconv.Itoa(random)

    altDataMap := make(map[string]string)
    altDataMap["alt"] = strconv.Itoa(random)
    // pushing alt value to node, encoded as string. every sent string/ alt value represents one element/ msg in the topic
    rcfNodeClient.TopicPublishGlobData(client, "radarsensmglob", radDataMap)
    rcfNodeClient.TopicPublishGlobData(client, "altsensmglob", altDataMap)

	// equal 10 & 100 Hz
    // time.Sleep(100*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
  }

  // closing node conn at program end
  rcfNodeClient.NodeCloseConn(client)
}