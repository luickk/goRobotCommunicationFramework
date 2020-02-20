package main

import (
 node_client "robot-communication-framework/rcf_cc_node_client"
 "time"
 "strconv"
)

func main() {

  node_client.Create_topic("test", 200)

  i := 0
  for range time.Tick(time.Second * 1) {
    i++
    node_client.Push_data("b-"+strconv.Itoa(i), 200, "test")
  }

  // node_client.List_cctopics(200, "test")
}
