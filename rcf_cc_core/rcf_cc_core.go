package rcf_cc_core

import (
  "fmt"
)

// initiating node with given id
func Init(node_id int) {
  fmt.Println("initiating node with name", node_id)
}

// reads one line (till null byte) from given command&control topic
func read_cctopic(node_id int) []byte{

  return []byte{22}
}

// write given data to  command&control topic
func publish_cctopic(data []byte, node_id int) {

}

// create command&control topic
func create_cctopic(node_id int) int {
  fmt.Println("creating topic on port, with id: ", node_id)

  return 1
}
