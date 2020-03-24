/*
Robot Communication Framework

 The RCF is a framework for data distribution, which the most essential part of an autonomous platform.
 It is very similar to ROS but without packages, the C/C++ complexity overhead while still maintaining speed and safe
 thread/ lang. standards, thanks to the go lang.

*/

package rcf_node

import (
  "fmt"
	"bufio"
  "bytes"
	"net"
  "time"
	"strings"
  "strconv"
  "robot-communication-framework/rcf_util"
)

// node msg/ element history length
var topic_capacity = 5

// frequency with which nodes handlers are regfreshed
var node_freq = 1000000

// node struct
type Node struct {
  // id or port of node
  id int

  // key: topic name, value: stack slice
  topics map[string][][]byte

  // channel map with first key(topic name) value(msg, topic element) pair, whichs element is then pushed to topic with topic name
  topic_push_ch chan map[string][]byte

  // channel map with first key(topic name) value(msg, topic element) pair, whichs element is then pushed to topic with topic name
  topic_create_ch chan string

  // channel map with first key(topic name) value(listening conn) pair, that's then added to listener conn
  topic_listener_conn_ch chan map[net.Conn]string

  // action map with first key(action name) value(anon action func) pair
  actions map[string]action_fn

  // action map with first key(action name) value(anon action func) that's then added to actions map
  action_create_ch chan map[string]action_fn

  // string channel with each string representing a action name executed when pushed to channel
  action_exec_ch chan string
}

// general action function type
type action_fn func (node_instance Node)

// handles every incoming node client connection
func handle_Connection(node Node, conn net.Conn) {
  defer conn.Close()

  for {
    data_b := make([]byte, 512)
    n, err_handle := bufio.NewReader(conn).Read(data_b)
    data_b = data_b[:n]
    data := string(data_b)
    if err_handle != nil {
      fmt.Println("/[node] ", err_handle)
      return
    }

    if len(data) > 0 {
      // fmt.Println(data)

      // literal commands wihtout args
      if rcf_util.Trim_suffix(data) == "end" {
        fmt.Println("/[conn]")
        conn.Close()
        return
      } else if rcf_util.Trim_suffix(data) =="list_topics" {
        conn.Write([]byte(strings.Join(Node_list_topics(node), ",")+"\n"))
      }

      // cmds with args/ require parsing
      push_rdata:=strings.Split(data, "+")
      pull_rdata:=strings.Split(data, "-")

      // data pushed to topic
      if len(push_rdata)>=2 && string(data[0])!="+" {
        topic_name := push_rdata[0]
        data_payload := data_b[len(push_rdata[0])+1:]
        Topic_publish_data(node, topic_name, data_payload)

      // data pulled from stack
      } else if len(pull_rdata) >=2 && string(data[0])!="+" {
        topic_name := pull_rdata[0]
        elements,_ := strconv.Atoi(rcf_util.Trim_suffix(pull_rdata[1]))
        data_b := Topic_pull_data(node, topic_name, elements)
        if(elements<=1) {
          conn.Write(append(data_b[0], '\r'))
        } else {
          tdata := bytes.Join(data_b, []byte("\r"))
          conn.Write(tdata)
        }
      } else if string(data[0])=="+" {
        Topic_create(node, data)

      // $ enables continuous data streaming mode, in whichthe topics data is continuously send to the client
      } else if string(data[0])=="$" {
        topic_name := rcf_util.Apply_naming_conv(data)
        Topic_add_listener_conn(node, topic_name, conn)

      } else if string(data[0])=="*" {
        exec_action_name := rcf_util.Apply_naming_conv(data)
        Action_exec(node, exec_action_name)
      }
      // fmt.Println(topics)
      data = ""
    }
  }
}

// handles all memory critical write operations to topic map and
// reduces the topics slice to given max length
func topic_handler(node Node) {
  listener_conns := make(map[net.Conn]string)
  for {
    select {
      case listener_topic_map := <-node.topic_listener_conn_ch:
        listening_conn := rcf_util.Get_first_map_key_cs(listener_topic_map)
        listener_conns[listening_conn] = listener_topic_map[listening_conn]
        fmt.Println("listener added")
        
      case topic_element_map := <-node.topic_push_ch:
        topic_name := rcf_util.Get_first_map_key_ss(topic_element_map)

        if rcf_util.Topics_contains_topic(node.topics, topic_name){
          topic_val_element := topic_element_map[topic_name]

          node.topics[topic_name] = append(node.topics[topic_name], topic_val_element)

          // check if topic exceeds topic cap limits
          if len(node.topics[topic_name]) > topic_capacity {
            topic_overhead := len(node.topics[topic_name])-topic_capacity
            // slicing size of slice to right size‚
            node.topics[topic_name] = node.topics[topic_name][topic_overhead:]
          }

          // check if topic, which data is pushed to, has a listening conn
          for k, v := range listener_conns {
            if v == topic_name {
              k.Write([]byte(topic_val_element))
            }
          }
        }

      case topic_create_name := <- node.topic_create_ch:
        fmt.Println("+[topic] ", topic_create_name)
        if rcf_util.Topics_contains_topic(node.topics, topic_create_name) {
          fmt.Println("/[topic] ", topic_create_name)
        } else {
          node.topics[topic_create_name] = [][]byte{}
        }
    }
    time.Sleep(time.Duration(node_freq))
  }
}

func action_handler(node_instance Node) {
  for {
    select {
    case create_action_map := <- node_instance.action_create_ch:
      var create_action_name string
      for k := range create_action_map { create_action_name = k }
      node_instance.actions[create_action_name] = create_action_map[create_action_name]
    case action_exec_ch := <-node_instance.action_exec_ch:
      action_func := node_instance.actions[action_exec_ch]
      go action_func(node_instance)
    }
    time.Sleep(time.Duration(node_freq))
  }
}



// creating node instance struct
func Create(node_id int) Node{
  // key: topic name, value: stack slice
  topics := make(map[string][][]byte)

  // channel map with first key(topic name) value(msg, topic element) pair, whichs element is then pushed to topic with topic name
  topic_push_ch := make(chan map[string][]byte)

  // channel map with first key(topic name) value(msg, topic element) pair, whichs element is then pushed to topic with topic name
  topic_create_ch := make(chan string)

  // channel map with first key(topic name) value(listening conn) pair, that's then added to listener conn
  topic_listener_conn_ch := make(chan map[net.Conn]string)

  // action map with first key(action name) value(anon action func) pair
  actions := make(map[string]action_fn)

  // action map with first key(action name) value(anon action func) that's then added to actions map
  action_create_ch := make(chan map[string]action_fn)

  // string channel with each string representing a action name executed when pushed to channel
  action_exec_ch := make(chan string)

  return Node{node_id, topics, topic_push_ch, topic_create_ch, topic_listener_conn_ch, actions, action_create_ch, action_exec_ch}
}

// createiating node with given id
// returns createiated node instance to enable direct action and topic operations
func Init(node Node) {
  fmt.Println("+[node] ", node.id)

  go topic_handler(node)

  go action_handler(node)

  var port string = ":"+strconv.Itoa(node.id)

  l, err := net.Listen("tcp4", port)

  if err != nil {
    fmt.Println("/[node] ", err)
  }

  defer l.Close()
  for {
    conn, err_handle := l.Accept()
    if err_handle != nil {
      fmt.Println("/[node] ",err_handle)
    }
    go handle_Connection(node, conn)
  }
}

func Node_halt() {
  for{time.Sleep(1*time.Second)}
}

func Topic_add_listener_conn(node Node, topic_name string, conn net.Conn) {
  topic_name = rcf_util.Apply_naming_conv(topic_name)
  fmt.Println("cpull ", topic_name)
  if rcf_util.Topics_contains_topic(node.topics, topic_name) {
    node.topic_listener_conn_ch <- map[net.Conn]string {conn: topic_name}
  }
}

func Node_list_topics(node Node) []string{
  fmt.Println("listing topics")
  keys := make([]string, 0, len(node.topics))
  for k, v := range node.topics {
    v=v
    keys = append(keys, k)
  }
  if len(node.topics) > 0 {
    return keys
  } else if len(node.topics) == 0 {
    return []string{"none"}
  }
  return []string{"none"}
}

func Topic_pull_data(node Node, topic_name string, elements int) [][]byte {
  if elements >= len(node.topics[topic_name]){
    return node.topics[topic_name]
  } else {
    return node.topics[topic_name][:elements]
  }
  return [][]byte{}
}

func Topic_publish_data(node Node, topic_name string, tdata []byte) {
  node.topic_push_ch <- map[string][]byte {topic_name: tdata}
  fmt.Println("->[topic] ", topic_name)
}

// create command&control topic
func Topic_create(node Node, topic_name string) {
  topic_name = rcf_util.Apply_naming_conv(topic_name)

  node.topic_create_ch <- topic_name
}

func Action_create(node Node, action_name string, action_func action_fn) {
    action_name = rcf_util.Apply_naming_conv(action_name)
    node.action_create_ch <- map[string]action_fn {action_name: action_func}
}

func Action_exec(node Node, action_name string) {
    node.action_exec_ch <- action_name
}
