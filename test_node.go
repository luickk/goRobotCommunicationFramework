package main

import (
  "fmt"
  rcf_node "robot-communication-framework/rcf_node"
)

func main() {
  topic_push_ch, topic_init_ch, service_init_ch := rcf_node.Init(28)

  rcf_node.Service_init(service_init_ch, "test", func(){
    fmt.Println("---- SERVICE TEST EXECUTED")
  })

  fmt.Println(topic_push_ch, topic_init_ch, service_init_ch)

}
