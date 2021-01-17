package main

import (
	"fmt"
  "os"

	"goRobotCommunicationFramework/rcfNodeClient"
)

func main() {
  topicNameArg := os.Args[1]
	// opening connection(tcp client) to node with id(port) 30
	client, err := rcfNodeClient.New(47)
  if err != nil {
    fmt.Println(err)
  }
  var res [][]byte
  for {
    // pulling last 2 msgs from topic
    res, err = client.TopicPullData(topicNameArg, 1)
		if err != nil {
			fmt.Println(err)
			return
		}
    fmt.Println("- Single glob pull results("+topicNameArg+"): ")
    fmt.Println(string(res[0]))
  }
}
