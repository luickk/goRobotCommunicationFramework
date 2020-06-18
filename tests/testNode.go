package main

import (
	"fmt"
	"strings"
	"time"
	"rcf/rcfNode"
)

func main() {
  node_instance := rcfNode.Create(28)

  rcfNode.Init(node_instance)

  rcfNode.Action_create(node_instance, "test", func(params []byte, n rcfNode.Node){
    fmt.Println("---- ACTION TEST EXECUTED. Active Topics: " + strings.Join(rcfNode.Node_list_topics(n), ","))
    fmt.Println("Params: " + string(params))
  })

  rcfNode.Action_create(node_instance, "test2", func(params []byte, n rcfNode.Node){
    fmt.Println("---- ACTION TEST EXECUTED. Active Topics: " + strings.Join(rcfNode.Node_list_topics(n), ","))
    fmt.Println("Params: " + string(params))
  })

  rcfNode.Service_create(node_instance, "0secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("---- SERVICE 0secproc EXECUTED")
    fmt.Println("Params: " + string(params))
    return []byte("result")
  })

  rcfNode.Service_create(node_instance, "3secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("---- SERVICE 3secproc EXECUTED")
    fmt.Println("Params: " + string(params))
    time.Sleep(3*time.Second)
    return []byte("result")
  })

  rcfNode.Service_create(node_instance, "10secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("---- SERVICE 10secproc EXECUTED")
    fmt.Println("Params: " + string(params))
    time.Sleep(10*time.Second)
    return []byte("result")
  })

  rcfNode.Node_halt()
}
