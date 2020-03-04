package main

import (
  "fmt"
  node_client "robot-communication-framework/rcf_cc_node_client"
)

func main() {
  conn := node_client.Connect_to_cc_node(30)

  topic_listener := node_client.Continuous_data_pull(conn, "altsens")

  for {
    select {
    case alt := <-topic_listener:
        fmt.Println("Altitude changed: ", alt)
    }
  }

  node_client.Close_cc_node(conn)
}
