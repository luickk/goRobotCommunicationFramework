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
	"net"
	"strings"
  "strconv"
  "robot-communication-framework/rcf_util"
)

var topic_capacity = 5

// node struct
type node struct {
  // id or port of node
  id int

  // key: topic name, value: stack slice
  topics map[string][]string

  // channel map with first key(topic name) value(msg, topic element) pair, whichs element is then pushed to topic with topic name
  topic_push_ch chan map[string]string

  // channel map with first key(topic name) value(msg, topic element) pair, whichs element is then pushed to topic with topic name
  topic_init_ch chan string

  // channel map with first key(topic name) value(listening conn) pair, that's then added to listener conn
  topic_listener_conn_ch chan map[net.Conn]string

  // service map with first key(service name) value(anon service func) pair
  services map[string]service_fn

  // service map with first key(service name) value(anon service func) that's then added to services map
  service_init_ch chan map[string]service_fn

  // string channel with each string representing a service name executed when pushed to channel
  service_exec_ch chan string
}

// general service function type
type service_fn func (node_instance node)

// handles every incoming node client connection
func handle_Connection(node node, conn net.Conn) {
  defer conn.Close()

  for {
    data, err_handle := bufio.NewReader(conn).ReadString('\n')

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
        tdata := push_rdata[1]
        Topic_push_data(node, topic_name, tdata)

      // data pulled from stack
      } else if len(pull_rdata) >=2 && string(data[0])!="+" {
        topic_name := pull_rdata[0]
        elements,_ := strconv.Atoi(rcf_util.Trim_suffix(pull_rdata[1]))
        conn.Write([]byte(strings.Join(Topic_pull_data(node, topic_name, elements), ",")+"\n"))

      } else if string(data[0])=="+" {
        Topic_init(node, data)

      // $ enables continuous data streaming mode, in whichthe topics data is continuously send to the client
      } else if string(data[0])=="$" {
        topic_name := rcf_util.Apply_naming_conv(data)
        Topic_add_listener_conn(node, topic_name, conn)
      } else if string(data[0])=="*" {
        exec_service_name := rcf_util.Apply_naming_conv(data)
        Service_exec(node, exec_service_name)
      }
      // fmt.Println(topics)
      data = ""
    }
  }
}

// handles all memory critical write operations to topic map and
// reduces the topics slice to given max length
func topic_handler(node node) {
  listener_conns := make(map[net.Conn]string)
  for {
    select {

    case listener_topic_map := <-node.topic_listener_conn_ch:
        listening_conn := rcf_util.Get_first_map_key_cs(listener_topic_map)
        listener_conns[listening_conn] = listener_topic_map[listening_conn]
      case topic_element_map := <-node.topic_push_ch:
        topic_name := rcf_util.Get_first_map_key_ss(topic_element_map)

        if rcf_util.Topics_contains_topic(node.topics, topic_name){
          topic_val_element := topic_element_map[topic_name]

          node.topics[topic_name] = append(node.topics[topic_name], topic_val_element)

          // check of topic exceeds topic cap limits
          if len(node.topics[topic_name]) > topic_capacity {
            topic_overhead := len(node.topics[topic_name])-topic_capacity
            // slicing size of slice to right size
            node.topics[topic_name] = node.topics[topic_name][topic_overhead:]
          }

          // check if topic, which data is pushed to, has a listening conn
          for k, v := range listener_conns {
            if v == topic_name {
              k.Write([]byte(topic_val_element+"\n"))
            }
          }
        }
      case topic_init_name := <- node.topic_init_ch:
        fmt.Println("+[topic] ", topic_init_name)
        if rcf_util.Topics_contains_topic(node.topics, topic_init_name) {
          fmt.Println("/[topic] ", topic_init_name)
        } else {
          node.topics[topic_init_name] = []string{"init"}
        }
    }
  }
}

func service_handler(node node) {
  for {
    select {
    case init_service_map := <- node.service_init_ch:
      var init_service_name string
      for k := range init_service_map { init_service_name = k }
      node.services[init_service_name] = init_service_map[init_service_name]
    case service_exec_ch := <-node.service_exec_ch:
      service_func := node.services[service_exec_ch]

      service_func()

    default:

    }
  }
}



// creating node instance struct
func Create(node_id int) node{
  // key: topic name, value: stack slice
  topics := make(map[string][]string)

  // channel map with first key(topic name) value(msg, topic element) pair, whichs element is then pushed to topic with topic name
  topic_push_ch := make(chan map[string]string)

  // channel map with first key(topic name) value(msg, topic element) pair, whichs element is then pushed to topic with topic name
  topic_init_ch := make(chan string)

  // channel map with first key(topic name) value(listening conn) pair, that's then added to listener conn
  topic_listener_conn_ch := make(chan map[net.Conn]string)

  // service map with first key(service name) value(anon service func) pair
  services := make(map[string]service_fn)

  // service map with first key(service name) value(anon service func) that's then added to services map
  service_init_ch := make(chan map[string]service_fn)

  // string channel with each string representing a service name executed when pushed to channel
  service_exec_ch := make(chan string)

  return node{node_id, topics, topic_push_ch, topic_init_ch, topic_listener_conn_ch, services, service_init_ch, service_exec_ch}
}

// initiating node with given id
// returns initiated node instance to enable direct service and topic operations
func Init(node node) {
  fmt.Println("+[node] ", node.id)

  go topic_handler(node)

  go service_handler(node)

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
  for{}
}

func Topic_add_listener_conn(node node, topic_name string, conn net.Conn) {
    topic_name = rcf_util.Apply_naming_conv(topic_name)
    fmt.Println("cpull ", topic_name)
    if rcf_util.Topics_contains_topic(node.topics, topic_name) {
      node.topic_listener_conn_ch <- map[net.Conn]string {conn: topic_name}
    }
}

func Node_list_topics(node node) []string{
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

func Topic_pull_data(node node, topic_name string, elements int) []string {
  if elements >= len(node.topics[topic_name]){
    return node.topics[topic_name]
  } else {
    return node.topics[topic_name][:elements]
  }
  return []string{"none"}
}

func Topic_push_data(node node, topic_name string, tdata string) {
  if rcf_util.Topics_contains_topic(node.topics, topic_name) {
    node.topic_push_ch <- map[string]string {topic_name: rcf_util.Trim_suffix(tdata)}
    fmt.Println("->[topic] ", topic_name)
  } else {
    fmt.Println("/+[topic] ", topic_name)
  }
}

// create command&control topic
func Topic_init(node node, topic_name string) {
  topic_name = rcf_util.Apply_naming_conv(topic_name)

  node.topic_init_ch <- topic_name
}

func Service_init(node node, service_name string, service_func service_fn) {
    service_name = rcf_util.Apply_naming_conv(service_name)
    node.service_init_ch <- map[string]service_fn {service_name: service_func}
}

func Service_exec(node node, service_name string) {
    node.service_exec_ch <- service_name
}
