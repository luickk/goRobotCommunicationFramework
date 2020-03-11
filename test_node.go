package main

import (
  "fmt"
  rcf_node "robot-communication-framework/rcf_node"
)

func main() {
  node := rcf_node.Create(28)

  go rcf_node.Init(node)

  rcf_node.Service_init(node, "test", func(){
    fmt.Println("---- SERVICE TEST EXECUTED")
  })

  rcf_node.Node_halt()

}
