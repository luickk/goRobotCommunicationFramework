package main

import (
  "fmt"
  "strings"
  "time"
  rcf_node "robot-communication-framework/rcf_node"
)

func main() {
  node_instance := rcf_node.Create(28)

  go rcf_node.Init(node_instance)

  rcf_node.Action_create(node_instance, "test", func(n rcf_node.Node){
    fmt.Println("---- ACTION TEST EXECUTED. Active Topics: " + strings.Join(rcf_node.Node_list_topics(n), ","))
  })

  rcf_node.Action_create(node_instance, "test2", func(n rcf_node.Node){
    fmt.Println("---- ACTION TEST EXECUTED. Active Topics: " + strings.Join(rcf_node.Node_list_topics(n), ","))
  })

  rcf_node.Service_create(node_instance, "0secproc", func(n rcf_node.Node) []byte{
    fmt.Println("---- SERVICE 0secproc EXECUTED")
    return []byte("result")
  })

  rcf_node.Service_create(node_instance, "3secproc", func(n rcf_node.Node) []byte{
    fmt.Println("---- SERVICE 3secproc EXECUTED")
    time.Sleep(3*time.Second)
    return []byte("result")
  })

  rcf_node.Service_create(node_instance, "10secproc", func(n rcf_node.Node) []byte{
    fmt.Println("---- SERVICE 10secproc EXECUTED")
    time.Sleep(10*time.Second)
    return []byte("result")
  })

  rcf_node.Node_halt()
}
