package main

import (
  "fmt"
  "strings"
  rcf_node "robot-communication-framework/rcf_node"
)

func main() {
  // creating node instance object which contains node struct in which all intern comm channels and topic/ service data maps are contained
  node_instance := rcf_node.Create(30)

  // initiating node by opening tcp server on node id
  // strarting service and topic handlers
  go rcf_node.Init(node_instance)

  // adding service
  rcf_node.Service_init(node_instance, "test", func(n rcf_node.Node){
    fmt.Println("---- SERVICE TEST EXECUTED. Active Topics: " + strings.Join(rcf_node.Node_list_topics(n), ","))

  })

  // halting node so it doesn't quit
  rcf_node.Node_halt()
}
