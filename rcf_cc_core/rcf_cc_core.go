package rcf_cc_core

import (
  "fmt"
)

func Init(name string){
  fmt.Println("creating node with name %d", name)
}

func create_topic(name string, port int) int {
  fmt.Println("creating topic on port %i with name %d", port, name)

  return 1
}
