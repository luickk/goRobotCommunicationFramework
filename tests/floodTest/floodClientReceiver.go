package main

import (
	"fmt"
  "os"

	"rcf/rcfNodeClient"
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
    res, err = client.TopicPullData(topicNameArg, 2)
		if err != nil {
			fmt.Println(err)
			return
		}
    fmt.Println("- floodClientReceiver: Single glob pull results: ")
    fmt.Println(res)
  }
}
