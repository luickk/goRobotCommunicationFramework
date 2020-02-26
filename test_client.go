package main

import (
  "os"
  node_client "robot-communication-framework/rcf_cc_node_client"
)

func main() {
  conn := node_client.Connect_to_cc_node(28)

  args := os.Args

  if args[1] == "ct" {
    node_client.Create_topic(conn, "test")
  } else if args[1] == "pt" {
    node_client.Push_data(conn, "b2", "test")
  }

  // node_client.Push_data(conn, "b1", "test")
  // node_client.Push_data(conn, "b1", "test")

  // fmt.Println(node_client.Pull_data(conn, 3, "test"))

  node_client.Close_cc_node(conn)

  conn.Close()
}
