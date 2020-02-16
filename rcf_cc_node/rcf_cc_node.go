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
  "robot-communication-framework/rcf_util"
)

// key: topic name, value: stack slice
var topics map[string][]string

// handles every incoming node client connection
func handle_Connection(conn net.Conn) {
  fmt.Println("Handling, ", conn.RemoteAddr().String())

  for {
    data, err_handle := bufio.NewReader(conn).ReadString('\n')
    if err_handle != nil {
      fmt.Println(err_handle)
      return
    }
    if data == "end" {
      break
    }

    push_rdata:=strings.Split(data, "+")
    pull_rdata:=strings.Split(data, "-")

    // data pushed t stack
    if len(push_rdata)>=2 && string(data[0])!="+" {
      topic := push_rdata[0]
      tdata := push_rdata[1]

      fmt.Println("Topic push request, to topic: ", topic)
      fmt.Println("Data: ", tdata)

      if val, ok := topics[topic]; ok {
        val = val
        fmt.Println("Added data to topic")
      } else {
        fmt.Println("Topic not found")
      }

    // data pulled/ poped fro stack
    } else if len(pull_rdata) >=2 && string(data[0])!="+" {
        topic := pull_rdata[0]
        elements := pull_rdata[1]

        fmt.Println("Topic pull/pop request, from topic: ", topic)
        fmt.Println("Elements: ", elements)

        if val, ok := topics[topic]; ok {
          val = val
          fmt.Println(elements," popped from topic")
        } else {
          fmt.Println("Topic not found")
        }

    } else if string(data[0])=="+" {
      Create_cctopic(data)
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

// create command&control topic
func Create_cctopic(topic_name string) {
  topic_name = rcf_util.Apply_naming_conv(topic_name)
  fmt.Println("creating topic, ", topic_name)

  topics := make(map[string][]string)
  topics[topic_name] = []string{"test elemet","s"}
  fmt.Println(topics)
}
