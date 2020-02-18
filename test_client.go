package main

import (
 node_client "robot-communication-framework/rcf_cc_node_client"
)

func main() {

  // node_client.Create_topic("test", 200)
  // node_client.Push_data("b", 200, "test")
  // node_client.Pop_element(200, "test")

  node_client.List_cctopics(200, "test")
}
