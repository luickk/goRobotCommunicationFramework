package main

import (
	"fmt"
	"strings"
	"time"
	"goRobotCommunicationFramework/rcfNode"
)

func main() {
	errorStream := make(chan error)

  node := rcfNode.New(8000, errorStream)

  node.ActionCreate("test", func(params []byte, n rcfNode.Node){
    fmt.Println("- action test executed, active Topics: " + strings.Join(node.NodeListTopics(), ","))
    fmt.Println("Params: " + string(params))
  })

  node.ServiceCreate("0secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("- service 0secproc executed")
    fmt.Println("Params: " + string(params))
    return []byte("result")
  })

  node.ServiceCreate("3secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("- service 3secproc executed")
    fmt.Println("Params: " + string(params))
    time.Sleep(3*time.Second)
    return []byte("result")
  })

  node.ServiceCreate("10secproc", func(params []byte, n rcfNode.Node) []byte{
    fmt.Println("- service 10secproc executed")
    fmt.Println("Params: " + string(params))
    time.Sleep(10*time.Second)
    return []byte("result")
  })

  for{
		if err :=<- errorStream; err != nil {
			fmt.Println(err.Error())
		}
	}
}
