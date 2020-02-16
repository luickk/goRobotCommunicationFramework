package main

import (
 node_client "robot-communication-framework/rcf_cc_node_client"
)

func main() {

  // node_client.Create_topic("test", 200)
  node_client.Push_data("dsadsad", 200, "test")
  // node_client.Pop_data(2, 200, "test")
}
