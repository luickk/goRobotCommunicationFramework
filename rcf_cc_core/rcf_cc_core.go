/*
Robot Communication Framework

 The RCF is a framework for data distribution, which the most essential part of an autonomous platform.
 It is very similar to ROS but without packages, the C/C++ complexity overhead while still maintaining speed and safe
 thread/ lang. standards, thanks to the go lang.

*/

package rcf_cc_core

import (
  "fmt"
	"bufio"
	"net"
	"strings"
  "strconv"
)

func handle_Connection(conn net.Conn) {
  fmt.Println("Handling, ", conn.RemoteAddr().String())

  for {
    data, err_handle := bufio.NewReader(conn).ReadString('\n')

    if err_handle != nil {
      fmt.Println(err_handle)

      return
    }

    temp := strings.TrimSpace(string(data))

    if temp == "end" {
      break
    }
  }
}

// initiating node with given id
func Init(node_id int) {
  fmt.Println("initiating node with name", node_id)

  var port string = ":"+strconv.Itoa(node_id)

  l, err := net.Listen("tcp4", port)

  if err != nil {
    fmt.Println("an error occured: ")
    fmt.Println(err)

    return
  }

  defer l.Close()

  for {
    conn, err_handle := l.Accept()
    if err_handle != nil {
      fmt.Println(err_handle)
      return
    }
    go handle_Connection(conn)
  }
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
