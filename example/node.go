package main

import (
  "fmt"
  rcf_node "robot-communication-framework/rcf_node"
)

func main() {
  // creating node instance object which contains node struct in which all intern comm channels and topic/ action data maps are contained
  node_instance := rcf_node.Create(30)

  // initiating node by opening tcp server on node id
  // strarting action and topic handlers
  go rcf_node.Init(node_instance)

  // adding action
  rcf_node.Action_create(node_instance, "test", func(n rcf_node.Node){
    fmt.Println("---- ACTION TEST EXECUTED.")
  })

  // halting node so it doesn't quit
  rcf_node.Node_halt()
}
