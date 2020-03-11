package main

import (
  "fmt"
  "strings"
  rcf_node "robot-communication-framework/rcf_node"
)

func main() {
  node_instance := rcf_node.Create(28)

  go rcf_node.Init(node_instance)

  rcf_node.Service_init(node_instance, "test", func(n rcf_node.Node){
    fmt.Println("---- SERVICE TEST EXECUTED. Active Topics: " + strings.Join(rcf_node.Node_list_topics(n), ","))

  })

  rcf_node.Node_halt()
}
