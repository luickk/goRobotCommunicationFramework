package main

import (
  node_client "robot-communication-framework/rcf_cc_node_client"
)

func main() {

  conn := node_client.Connect_to_cc_node(28)

  // node_client.Create_topic(conn, "test")

  node_client.Push_data(conn, "b1", "test")
  // node_client.Push_data(conn, "b1", "test")
  // node_client.Push_data(conn, "b1", "test")

  // fmt.Println(node_client.Pull_data(conn, 3, "test"))
}
