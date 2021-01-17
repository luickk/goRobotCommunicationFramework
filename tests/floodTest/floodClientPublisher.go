package main

import (
  "fmt"
  "strconv"
  "os"
  "math/rand"

	"rcf/rcfNodeClient"
)

func main() {
	// opening connection(tcp client) to node with id(port) 30
	client, err := rcfNodeClient.New(47)
  if err != nil {
    fmt.Println(err)
  }
  topicNameArg := os.Args[1]

	// creating topic by sending cmd to node
	client.TopicCreate(topicNameArg)

  var alt int

	// loop to create sample data which is pushed to topic
	for {
		// generating random int
		alt = rand.Intn(200)

		// pushing alt value to node, encoded as string. every sent string/ alt value represents one element/ msg in the topic
		client.TopicPublishData(topicNameArg, []byte(strconv.Itoa(alt)))
	}
}
