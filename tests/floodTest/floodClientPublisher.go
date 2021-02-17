package main

import (
  "fmt"
  "strconv"
  "os"
  "math/rand"

	"goRobotCommunicationFramework/rcfNodeClient"
)

func main() {
	// opening connection(tcp client) to node with id(port) 30
	errorStream := make(chan error)
  client, err := rcfNodeClient.New(8000, errorStream)
	if err != nil {
		fmt.Println(err)
		return
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
