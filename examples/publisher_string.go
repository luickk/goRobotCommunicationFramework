package main

import (
  "fmt"
  "math/rand"
  "strconv"
  node_client "robot-communication-framework/rcf_node_client"
)

func main() {
  // opening connection(tcp client) to node with id(port) 30
  conn := node_client.Node_open_conn(30)

  // creating topic by sending cmd to node
  node_client.Topic_create(conn, "altsens2")

  // loop to create sample data which is pushed to topic
  for {
    // generating random int
    alt := rand.Intn(100)
    // printing sample data
    fmt.Println(alt)
    // pushing alt value to node, encoded as string. every sent string/ alt value represents one element/ msg in the topic
    node_client.Topic_publish_data(conn, "altsens2", strconv.Itoa(alt))
  }

  // closing node conn at program end
  node_client.Node_close_conn(conn)
}
