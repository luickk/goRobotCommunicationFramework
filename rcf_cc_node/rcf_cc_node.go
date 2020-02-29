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

var topic_capacity = 5

// handles every incoming node client connection
func handle_Connection(push_ch chan <- map[string]string, conn net.Conn, topics map[string][]string) {
  defer conn.Close()

  for {
    data, err_handle := bufio.NewReader(conn).ReadString('\n')

    if err_handle != nil {
      fmt.Println("/ ", err_handle)
      return
    }

    if len(data) > 0 {
      // literal commands wihtout args
      if rcf_util.Trim_suffix(data) == "end" {
        fmt.Println("/[conn]")
        conn.Close()
        return
      } else if rcf_util.Trim_suffix(data) =="list_cctopics" {
        fmt.Println("listing topics")
        keys := make([]string, 0, len(topics))
        for k, v := range topics {
          v=v
        	keys = append(keys, k)
        }
        if len(topics) > 0 {
          conn.Write([]byte(strings.Join(keys, ",")+"\n"))
        } else if len(topics) == 0 {
          conn.Write([]byte("none\n"))
        }
      }

      // cmds with args/ require parsing
      push_rdata:=strings.Split(data, "+")
      pull_rdata:=strings.Split(data, "-")

      // data pushed to stack
      if len(push_rdata)>=2 && string(data[0])!="+" {
        topic := push_rdata[0]
        tdata := push_rdata[1]

        if val, ok := topics[topic]; ok {
          val = val
          push_ch <- map[string]string {topic: rcf_util.Trim_suffix(tdata)}
          fmt.Println("->[topic] ", topic)
        } else {
          fmt.Println("/+[topic] ", topic)
        }

      // data pulled from stack
      } else if len(pull_rdata) >=2 && string(data[0])!="+" {
        topic := pull_rdata[0]
        elements,_ := strconv.Atoi(rcf_util.Trim_suffix(pull_rdata[1]))
        if elements >= len(topics[topic]){
          conn.Write([]byte(strings.Join(topics[topic], ",")+"\n"))
        } else {
        conn.Write([]byte(strings.Join(topics[topic][:elements], ",")+"\n"))
        }
      } else if string(data[0])=="+" {
        Create_topic(data, topics)

      // $ enables continuous data streaming mode, in whichthe topics data is continuously send to the client
      } else if string(data[0])=="$" {
        topic_name := rcf_util.Apply_naming_conv(string(data[0]))
        if val, ok := topics[topic_name]; ok {
          val = val
          for {
            conn.Write([]byte(strings.Join(topics[topic_name], ",")))
          }
        }
      }
      fmt.Println(topics)
      data = ""
    }
  }
}

// handles all memory critical write operations to topic map and
// reduces the topics slice to given max length
func topic_handler(push_ch <- chan map[string]string, topics map[string][]string, topic_capacity int) {
  for {
    topic_element := <-push_ch
    topic_name := rcf_util.Get_first_map_key(topic_element)
    topic_val_element := topic_element[topic_name]

    topics[topic_name] = append(topics[topic_name], topic_val_element)

    if len(topics[topic_name]) > topic_capacity {
      topic_overhead := len(topics[topic_name])-topic_capacity
      // slicing size of slice to right size
      topics[topic_name] = topics[topic_name][topic_overhead:]
    }
  }
}

// initiating node with given id
func Init(node_id int) {
  fmt.Println("+[node] ", node_id)

  // key: topic name, value: stack slice
  topics := make(map[string][]string)

  push_ch := make(chan map[string]string)

  go topic_handler(push_ch, topics, topic_capacity)

  var port string = ":"+strconv.Itoa(node_id)

  l, err := net.Listen("tcp4", port)

  if err != nil {
    fmt.Println("/[node] ", err)
    return
  }

  defer l.Close()

  for {
    conn, err_handle := l.Accept()
    if err_handle != nil {
      fmt.Println("/[node] ",err_handle)
      return
    }
    go handle_Connection(push_ch, conn, topics)
  }
}

// create command&control topic
func Create_topic(topic_name string, topics map[string][]string) {
  topic_name = rcf_util.Apply_naming_conv(topic_name)
  fmt.Println("+[topic] ", topic_name)
  if val, ok := topics[topic_name]; ok {
    val = val
    fmt.Println("/[topic] ", topic_name)
  } else {
    topics[topic_name] = []string{"init"}
  }
}
