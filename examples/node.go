package main

import (
	"fmt"
	rcf_node "rcf/rcf-node"
)

func main() {
  // creating node instance object which contains node struct in which all intern comm channels and topic/ action data maps are contained
  node_instance := rcf_node.Create(47)

  // initiating node by opening tcp server on node id
  // strarting action and topic handlers
  go rcf_node.Init(node_instance)

  // adding action
  rcf_node.Action_create(node_instance, "testAction", func(params []byte, n rcf_node.Node){
    fmt.Println("---- ACTION TEST EXECUTED.")
  })

  // adding action
  rcf_node.Service_create(node_instance, "testService", func(params []byte, n rcf_node.Node) []byte {
    fmt.Println("---- Service TEST EXECUTED.")
    return []byte("done")
  })

  // halting node so it doesn't quit
  rcf_node.Node_halt()
}
