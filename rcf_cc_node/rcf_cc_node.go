/*
Robot Communication Framework

 The RCF is a framework for data distribution, which the most essential part of an autonomous platform.
 It is very similar to ROS but without packages, the C/C++ complexity overhead while still maintaining speed and safe
 thread/ lang. standards, thanks to the go lang.

*/

package rcf_cc_node

import (
  "fmt"
	"bufio"
	"net"
	"strings"
  "strconv"
)

var m map[string]int

var node_topics = make([][]string, 10)

func handle_Connection(conn net.Conn) {
  fmt.Println("Handling, ", conn.RemoteAddr().String())

  for {
    data, err_handle := bufio.NewReader(conn).ReadString('\n')
    if err_handle != nil {
      fmt.Println(err_handle)
      return
    }
    // rdata := strings.TrimSpace(string(data))
    if data == "end" {
      break
    }

    fmt.Println(data)

    push_rdata:=strings.Split(data, "+")
    pull_rdata:=strings.Split(data, "-")

    fmt.Println(len(push_rdata))
    fmt.Println(len(pull_rdata))



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

// create command&control topic
func create_cctopic(node_id int) int {
  fmt.Println("creating topic on port, with id: ", node_id)

  return 1
}
