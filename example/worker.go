package main

import (
  "fmt"
  node_client "robot-communication-framework/rcf_node_client"
)

func main() {
  conn := node_client.Node_open_conn(30)

  topic_listener := node_client.Topic_listener(conn, "altsens")

  for {
    select {
    case alt := <-topic_listener:
        fmt.Println("Altitude changed: ", alt)
    }
  }

  node_client.Node_close_conn(conn)
}
