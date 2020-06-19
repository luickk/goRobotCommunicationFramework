package main

import (
	"fmt"
	"strings"
	"time"
	rcfNode "rcf/rcfNode"
)

func main() {
  node := rcfNode.Create(28)
  
  rcfNode.Init(node)

  rcfNode.ActionCreate(node, "test", func(params []byte, n rcfNode.Node){
    fmt.Println("---- ACTION TEST EXECUTED. Active Topics: " + strings.Join(rcfNode.NodeListTopics(n), ","))
    fmt.Println("Params: " + string(params))
  })

  rcfNode.ActionCreate(node, "test2", func(params []byte, n rcfNode.Node){
    fmt.Println("---- ACTION TEST EXECUTED. Active Topics: " + strings.Join(rcfNode.NodeListTopics(n), ","))
    fmt.Println("Params: " + string(params))
  })

  rcfNode.ServiceCreate(node, "0secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("---- SERVICE 0secproc EXECUTED")
    fmt.Println("Params: " + string(params))
    return []byte("result")
  })

  rcfNode.ServiceCreate(node, "3secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("---- SERVICE 3secproc EXECUTED")
    fmt.Println("Params: " + string(params))
    time.Sleep(3*time.Second)
    return []byte("result")
  })

  rcfNode.ServiceCreate(node, "10secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("---- SERVICE 10secproc EXECUTED")
    fmt.Println("Params: " + string(params))
    time.Sleep(10*time.Second)
    return []byte("result")
  })

  rcfNode.NodeHalt()
}
