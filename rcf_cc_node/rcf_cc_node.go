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
  "time"
)

var topic_capacity = 5

// handles every incoming node client connection
func handle_Connection(conn net.Conn, topics map[string][]string) {
  defer conn.Close()
  fmt.Println("client connected")

  for {
    data, err_handle := bufio.NewReader(conn).ReadString('\n')
    if err_handle != nil {
      fmt.Println("Err: ", err_handle)
      return
    }
    fmt.Println(len(data))

    time.Sleep(2 * time.Second)
    if len(data) > 0 {
      push_rdata:=strings.Split(data, "+")
      pull_rdata:=strings.Split(data, "-")

      // data pushed t stack
      if len(push_rdata)>=2 && string(data[0])!="+" {
        topic := push_rdata[0]
        tdata := push_rdata[1]

        fmt.Println(topics)
        fmt.Println("Topic push request, to topic: ", topic)
        fmt.Println("Data: ", tdata)

        if val, ok := topics[topic]; ok {
          val = val
          fmt.Println("Added data to topic")
          topics[topic] = append(topics[topic], strings.TrimSuffix(push_rdata[1], "\n"))
          fmt.Println(topics)
        } else {
          fmt.Println("Topic not found")
        }

      // data pulled from stack
      } else if len(pull_rdata) >=2 && string(data[0])!="+" {
          topic := pull_rdata[0]
          elements,_ := strconv.Atoi(strings.TrimSuffix(pull_rdata[1], "\n"))

          fmt.Println("Topic pull request, from topic: ", topic)
          fmt.Println("Elements: ", elements)

          conn.Write([]byte(strings.Join(topics[topic][:elements], ",")+"\n"))


      } else if string(data[0])=="+" {
        Create_cctopic(data, topics)
      } else if data=="list_cctopics" {
        List_cctopics(topics)
      }
    }
  }
}

// reduces the topics slice to given max length
func topic_size_handler(topics map[string][]string, topic_capacity int) {
  for {
    for k, v := range topics {
      if len(v) > topic_capacity {
        topic_overhead := len(v)-topic_capacity
        // slicing size of slice to right size
        topics[k] = v[topic_overhead:]
      }
      // fmt.Println(len(v),"-",v)
    }
  }
}

// initiating node with given id
func Init(node_id int) {
  fmt.Println("initiating node with name", node_id)

  // key: topic name, value: stack slice
  topics := make(map[string][]string)

  go topic_size_handler(topics, topic_capacity)

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
      fmt.Println("node err: ",err_handle)
      return
    }
    go handle_Connection(conn, topics)
  }
}

// prints topic map
func List_cctopics(topics map[string][]string) {
  fmt.Println("Topics with elements: ")
  fmt.Println(topics)
}

// create command&control topic
func Create_cctopic(topic_name string, topics map[string][]string) {
  topic_name = rcf_util.Apply_naming_conv(topic_name)
  fmt.Println("creating topic, ", topic_name)
  if val, ok := topics[topic_name]; ok {
    fmt.Println(topic_name, "-", val, " already exists")
  } else {
    topics[topic_name] = []string{"init"}
    fmt.Println(topics)
  }
}
