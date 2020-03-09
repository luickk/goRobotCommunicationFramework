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

// general service function type
type service_fn func ()

// handles every incoming node client connection
func handle_Connection(topic_push_ch chan <- map[string]string, topic_init_ch chan string, topic_listener_conn_ch chan <- map[net.Conn]string, service_exec_ch chan string, conn net.Conn, topics map[string][]string) {
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
        conn.Write([]byte(strings.Join(Node_list_topics(topics), ",")+"\n"))
      }

      // cmds with args/ require parsing
      push_rdata:=strings.Split(data, "+")
      pull_rdata:=strings.Split(data, "-")

      // data pushed to topic
      if len(push_rdata)>=2 && string(data[0])!="+" {
        topic_name := push_rdata[0]
        tdata := push_rdata[1]
        Topic_push_data(topic_push_ch, topics, topic_name, tdata)

      // data pulled from stack
      } else if len(pull_rdata) >=2 && string(data[0])!="+" {
        topic_name := pull_rdata[0]
        elements,_ := strconv.Atoi(rcf_util.Trim_suffix(pull_rdata[1]))
        conn.Write([]byte(strings.Join(Topic_pull_data(topics, topic_name, elements), ",")+"\n"))

      } else if string(data[0])=="+" {
        Topic_init(topic_init_ch, data)

      // $ enables continuous data streaming mode, in whichthe topics data is continuously send to the client
      } else if string(data[0])=="$" {
        topic_name := rcf_util.Apply_naming_conv(data)
        Topic_add_listener_conn(topics, topic_listener_conn_ch, topic_name, conn)
      } else if string(data[0])=="*" {
        exec_service_name := rcf_util.Apply_naming_conv(data)
        Service_exec(service_exec_ch, exec_service_name)
      }
      // fmt.Println(topics)
      data = ""
    }
  }
}

// handles all memory critical write operations to topic map and
// reduces the topics slice to given max length
func topic_handler(topic_push_ch <- chan map[string]string, topic_init_ch <- chan string, topic_listener_conn_ch <- chan map[net.Conn]string, topics map[string][]string, topic_capacity int) {
  listener_conns := make(map[net.Conn]string)
  for {
    select {

      case listener_topic_map := <-topic_listener_conn_ch:
        listening_conn := rcf_util.Get_first_map_key_cs(listener_topic_map)
        listener_conns[listening_conn] = listener_topic_map[listening_conn]
      case topic_element_map := <-topic_push_ch:
        topic_name := rcf_util.Get_first_map_key_ss(topic_element_map)

        if rcf_util.Topics_contains_topic(topics, topic_name){
          topic_val_element := topic_element_map[topic_name]

          topics[topic_name] = append(topics[topic_name], topic_val_element)

          // check of topic exceeds topic cap limits
          if len(topics[topic_name]) > topic_capacity {
            topic_overhead := len(topics[topic_name])-topic_capacity
            // slicing size of slice to right size
            topics[topic_name] = topics[topic_name][topic_overhead:]
          }

          // check if topic, which data is pushed to, has a listening conn
          for k, v := range listener_conns {
            if v == topic_name {
              k.Write([]byte(topic_val_element+"\n"))
            }
          }
        }
      case topic_init_name := <-topic_init_ch:
        fmt.Println("+[topic] ", topic_init_name)
        if rcf_util.Topics_contains_topic(topics, topic_init_name) {
          fmt.Println("/[topic] ", topic_init_name)
        } else {
          topics[topic_init_name] = []string{"init"}
        }
    }
  }
}

func service_handler(service_init_ch <- chan map[string]service_fn, service_exec_ch chan string, services map[string]service_fn) {
  for {
    select {
    case init_service_map := <-service_init_ch:
      var init_service_name string
      for k := range init_service_map { init_service_name = k }
      services[init_service_name] = init_service_map[init_service_name]
    case service_exec_ch := <-service_exec_ch:
      services[service_exec_ch]()
    default:

    }
  }
}

// initiating node with given id
// returns topic init, push channel and service init, execute channel to enable direct service and topic operations
func Init(node_id int) (chan map[string]string, chan string, chan map[string]service_fn, chan string){
  fmt.Println("+[node] ", node_id)

  // key: topic name, value: stack slice
  topics := make(map[string][]string)

  // channel map with first key(topic name) value(msg, topic element) pair, whichs element is then pushed to topic with topic name
  topic_push_ch := make(chan map[string]string)

  // channel map with first key(topic name) value(msg, topic element) pair, whichs element is then pushed to topic with topic name
  topic_init_ch := make(chan string)

  // channel map with first key(topic name) value(listening conn) pair, that's then added to listener conn
  topic_listener_conn_ch := make(chan map[net.Conn]string)

  go topic_handler(topic_push_ch, topic_init_ch, topic_listener_conn_ch, topics, topic_capacity)

  services := make(map[string]service_fn)

  service_init_ch := make(chan map[string]service_fn)

  service_exec_ch := make(chan string)

  go service_handler(service_init_ch, service_exec_ch, services)

  var port string = ":"+strconv.Itoa(node_id)

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
    go handle_Connection(topic_push_ch, topic_init_ch, topic_listener_conn_ch, service_exec_ch, conn, topics)
  }
  return topic_push_ch, topic_init_ch, service_init_ch, service_exec_ch
}

func Topic_add_listener_conn(topics map[string][]string, topic_listener_conn_ch chan <- map[net.Conn]string, topic_name string, conn net.Conn) {
    topic_name = rcf_util.Apply_naming_conv(topic_name)
    fmt.Println("cpull ", topic_name)
    if rcf_util.Topics_contains_topic(topics, topic_name) {
      topic_listener_conn_ch <- map[net.Conn]string {conn: topic_name}
    }
}

func Node_list_topics(topics map[string][]string) []string{
  fmt.Println("listing topics")
  keys := make([]string, 0, len(topics))
  for k, v := range topics {
    v=v
    keys = append(keys, k)
  }
  if len(topics) > 0 {
    return keys
  } else if len(topics) == 0 {
    return []string{"none"}
  }
  return []string{"none"}
}

func Topic_pull_data(topics map[string][]string, topic_name string, elements int) []string {
  if elements >= len(topics[topic_name]){
    return topics[topic_name]
  } else {
    return topics[topic_name][:elements]
  }
  return []string{"none"}
}

func Topic_push_data(topic_push_ch chan<- map[string]string, topics map[string][]string, topic_name string, tdata string) {
  if rcf_util.Topics_contains_topic(topics, topic_name) {
    topic_push_ch <- map[string]string {topic_name: rcf_util.Trim_suffix(tdata)}
    fmt.Println("->[topic] ", topic_name)
  } else {
    fmt.Println("/+[topic] ", topic_name)
  }
}

// create command&control topic
func Topic_init(topic_init_ch chan string, topic_name string) {
  topic_name = rcf_util.Apply_naming_conv(topic_name)

  topic_init_ch <- topic_name
}

func Service_init(service_init_ch chan map[string]service_fn, service_name string, service_func service_fn) {
    service_name = rcf_util.Apply_naming_conv(service_name)
    service_init_ch <- map[string]service_fn {service_name: service_func}
}

func Service_exec(service_exec_ch chan string, service_name string) {
    service_exec_ch <- service_name
}
