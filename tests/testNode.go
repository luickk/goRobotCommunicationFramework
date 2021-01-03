package main

import (
	"fmt"
	"strings"
	"time"
	"rcf/rcfNode"
)

func main() {
  node := rcfNode.New(47)

  node.ActionCreate("test", func(params []byte, n rcfNode.Node){
    fmt.Println("---- ACTION TEST EXECUTED. Active Topics: " + strings.Join(rcfNode.NodeListTopics(n), ","))
    fmt.Println("Params: " + string(params))
  })

  node.ActionCreate("test2", func(params []byte, n rcfNode.Node){
    fmt.Println("---- ACTION TEST EXECUTED. Active Topics: " + strings.Join(rcfNode.NodeListTopics(n), ","))
    fmt.Println("Params: " + string(params))
  })

  node.ServiceCreate("0secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("---- SERVICE 0secproc EXECUTED")
    fmt.Println("Params: " + string(params))
    return []byte("result")
  })

  node.ServiceCreate("3secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("---- SERVICE 3secproc EXECUTED")
    fmt.Println("Params: " + string(params))
    time.Sleep(3*time.Second)
    return []byte("result")
  })

  node.ServiceCreate("10secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("---- SERVICE 10secproc EXECUTED")
    fmt.Println("Params: " + string(params))
    time.Sleep(10*time.Second)
    return []byte("result")
  })

  for{}
}
