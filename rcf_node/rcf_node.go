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
var node_freq = 0

var tcp_conn_buffer = 1024

type topic_msg struct {
  topic_name string
  msg []byte
}

type topic_listener_conn struct {
  listening_conn net.Conn
  topic_name string
}

type action struct {
  action_name string
  action_function action_fn
}

type service struct {
  service_name string
  service_function service_fn
}

type service_exec struct {
  service_name string
  service_call_conn net.Conn
}
// node struct
type Node struct {
  // id or port of node
  id int

  // key: topic name, value: stack slice
  topics map[string][][]byte

  topic_push_ch chan topic_msg

  topic_create_ch chan string

  topic_listener_conn_ch chan topic_listener_conn

  topic_listener_conns []topic_listener_conn

  // action map with first key(action name) value(anon type action func) pair
  actions map[string]action_fn

  action_create_ch chan action

  action_exec_ch chan string


  // service map with first key(service name) value(anon type services func) pair
  services map[string]service_fn

  service_create_ch chan service

  service_exec_ch chan service_exec
}

// general action function type
type action_fn func (node_instance Node)

// general service function type
type service_fn func (node_instance Node) []byte


// handles every incoming node client connection
func handle_Connection(node Node, conn net.Conn) {
  defer conn.Close()

  for {
    data_b := make([]byte, tcp_conn_buffer)
    n, err_handle := bufio.NewReader(conn).Read(data_b)
    data_b = data_b[:n]
    data := string(data_b)
    split_data_b := bytes.Split(data_b, []byte("\r"))
    split_data := strings.Split(data, "\r")

    // iterating ovre conn read buffer array, split by backslash r
    for i, data_b := range split_data_b {
      data = split_data[i]

      if err_handle != nil {
        fmt.Println("/[node] ", err_handle)
        return
      }

      if len(data) > 0 {
        // fmt.Println(data)

        // literal commands wihtout args
        if data == "end" {
          fmt.Println("/[conn]")
          conn.Close()
          return
        } else if data =="list_topics" {
		      // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>"
          conn.Write(append([]byte(">info-list_topics-1-"),[]byte(strings.Join(Node_list_topics(node), ",")+"\r")...))
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
          elements,_ := strconv.Atoi(pull_rdata[1])
          data_b := Topic_pull_data(node, topic_name, elements)
          if(elements<=1) {
			      // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>"
            if len(data_b) >= 1 {
              conn.Write(append(append([]byte(">topic-"+topic_name+"-1-"), data_b[0]...), []byte("\r")...))
            } else {
              conn.Write(append([]byte(">topic-"+topic_name+"-1-"), []byte("\r")...))
            }
          } else {
            if len(data_b) >= 1 {
              tdata := append(bytes.Join(data_b, []byte("\nm")), []byte("\r")...)
  			      // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
              conn.Write(append([]byte(">topic-"+topic_name+"-"+strconv.Itoa(elements)+"-"), tdata...))
            } else {
              conn.Write(append([]byte(">topic-"+topic_name+"-1-"), []byte("\r")...))
            }
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
        } else if string(data[0])=="#" {
          exec_action_name := rcf_util.Apply_naming_conv(data)
          Service_exec(node, conn, exec_action_name)
        }
        // fmt.Println(topics)
        data = ""
      }
    }
  }
}

// handles all memory critical write operations to topic map and
// reduces the topics slice to given max length
func topic_handler(node Node) {
  for {
    select {
    case topic_listener := <-node.topic_listener_conn_ch:
        node.topic_listener_conns = append(node.topic_listener_conns, topic_listener)

    case topic_msg := <-node.topic_push_ch:

      if rcf_util.Topics_contain_topic(node.topics, topic_msg.topic_name){

        node.topics[topic_msg.topic_name] = append(node.topics[topic_msg.topic_name], topic_msg.msg)

        // check if topic exceeds topic cap limits
        if len(node.topics[topic_msg.topic_name]) > topic_capacity {
          topic_overhead := len(node.topics[topic_msg.topic_name])-topic_capacity
          // slicing size of slice to right size‚
          node.topics[topic_msg.topic_name] = node.topics[topic_msg.topic_name][topic_overhead:]
        }

        // check if topic, which data is pushed to, has a listening conn
        for _,topic_listener := range node.topic_listener_conns {
          if topic_listener.topic_name == topic_msg.topic_name {
			      // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
            topic_listener.listening_conn.Write(append(append([]byte(">topic-"+topic_msg.topic_name+"-1-"),[]byte(topic_msg.msg)...), []byte("\r")...))
          }
        }
      }

      case topic_create_name := <- node.topic_create_ch:
        fmt.Println("+[topic] ", topic_create_name)
        if rcf_util.Topics_contain_topic(node.topics, topic_create_name) {
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
    case action := <- node_instance.action_create_ch:
      node_instance.actions[action.action_name] = action.action_function
    case action_exec := <- node_instance.action_exec_ch:
      if _, ok := node_instance.actions[action_exec]; ok {
        action_func := node_instance.actions[action_exec]
        go action_func(node_instance)
      } else {
        fmt.Println("/[action] ", action_exec)
      }
    }
    time.Sleep(time.Duration(node_freq))
  }
}

func service_handler(node_instance Node) {
  for {
    select {
      case service := <- node_instance.service_create_ch:
        node_instance.services[service.service_name] = service.service_function
      case service_exec := <-node_instance.service_exec_ch:
        if _, ok := node_instance.services[service_exec.service_name]; ok {
          go func() {
            service_result := append(node_instance.services[service_exec.service_name](node_instance), []byte("\r")...)

			// client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>"
            service_exec.service_call_conn.Write(append([]byte(">service-"+service_exec.service_name+"-1-"), service_result...))
          }()
        } else {
          fmt.Println("/[service] ", service_exec.service_name)
		      // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>"
          service_exec.service_call_conn.Write(append([]byte(">service-"+service_exec.service_name+"-1-"), []byte(service_exec.service_name+" not found \r")...))
        }
      time.Sleep(time.Duration(node_freq))
    }
  }
}



// creating node instance struct
func Create(node_id int) Node{
  // key: topic name, value: stack slice
  topics := make(map[string][][]byte)

  topic_push_ch := make(chan topic_msg)

  topic_create_ch := make(chan string)

  topic_listener_conn_ch := make(chan topic_listener_conn)

  topic_listener_conns := make([]topic_listener_conn,0)

  // action map with first key(action name) value(anon action func) pair
  actions := make(map[string]action_fn)

  action_create_ch := make(chan action)

  action_exec_ch := make(chan string)

  services := make(map[string]service_fn)

  service_create_ch := make(chan service)

  service_exec_ch := make(chan service_exec)

  return Node{node_id, topics, topic_push_ch, topic_create_ch, topic_listener_conn_ch, topic_listener_conns, actions, action_create_ch, action_exec_ch, services, service_create_ch, service_exec_ch}
}

// createiating node with given id
// returns createiated node instance to enable direct service and topic operations
func Init(node Node) {
  fmt.Println("+[node] ", node.id)

  go topic_handler(node)

  go action_handler(node)

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
  for{time.Sleep(1*time.Second)}
}

func Topic_add_listener_conn(node Node, topic_name string, conn net.Conn) {
  topic_name = rcf_util.Apply_naming_conv(topic_name)
  fmt.Println("cpull ", topic_name)
  topic_listener_conn := new(topic_listener_conn)
  topic_listener_conn.topic_name = topic_name
  topic_listener_conn.listening_conn = conn
  node.topic_listener_conn_ch <- *topic_listener_conn
  topic_listener_conn = nil
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
  topic_msg := new(topic_msg)
  topic_msg.topic_name = topic_name
  topic_msg.msg = tdata
  node.topic_push_ch <- *topic_msg
  fmt.Println("->[topic] ", topic_name)
  topic_msg = nil
}

// create command&control topic
func Topic_create(node Node, topic_name string) {
  topic_name = rcf_util.Apply_naming_conv(topic_name)

  node.topic_create_ch <- topic_name
}

func Action_create(node Node, action_name string, action_func action_fn) {
    action_name = rcf_util.Apply_naming_conv(action_name)
    new_action := new(action)
    new_action.action_name = action_name
    new_action.action_function = action_func
    node.action_create_ch <- *new_action
    new_action = nil
}

func Action_exec(node Node, action_name string) {
    node.action_exec_ch <- action_name
}

func Service_create(node Node, service_name string, service_func service_fn) {
    service_name = rcf_util.Apply_naming_conv(service_name)
    service := new(service)
    service.service_name = service_name
    service.service_function = service_func
    node.service_create_ch <- *service
    service = nil
}

func Service_exec(node Node, conn net.Conn, service_name string) {
  service_exec := new(service_exec)
  service_exec.service_name = service_name
  service_exec.service_call_conn = conn
  node.service_exec_ch <- *service_exec
  service_exec = nil
}
