package rcf_node_client

import(
	"fmt"
	"net"
	"strconv"
	"strings"
	"bufio"
	"bytes"
	"rcf/rcf-util"
)

var tcp_conn_buffer = 1024

// function to connect to tcp server (node) and returns connection
func connect_to_tcp_server(port int) net.Conn{
  conn, err := net.Dial("tcp4", ":"+strconv.Itoa(port))

  if err != nil {
    fmt.Println(err)
  }
  // don't forget to close connection
  return conn
}

// returns connection to node
func Node_open_conn(node_id int) net.Conn {
  var conn net.Conn = connect_to_tcp_server(node_id)

  return conn
}

func Node_close_conn(conn net.Conn) {
  conn.Write([]byte("end\r"))
  conn.Close()
}



// pushes raw byte slice msg to topic msg stack
func Topic_publish_raw_data(conn net.Conn, topic_name string, data []byte) {
  send_slice := append(append([]byte(">topic-"+topic_name+"-publish-"),data...),"\r"...)
  conn.Write(send_slice)
}

// pulls x msgs from topic topic stack
func Topic_pull_raw_data(conn net.Conn, nmsgs int, topic_name string) [][]byte {
  send_slice := append([]byte(">topic-"+topic_name+"-pull-"+strconv.Itoa(nmsgs)), "\r"...)
  conn.Write(send_slice)
  var msgs [][]byte
  rdata := make([]byte, tcp_conn_buffer)
  n, err_handle := bufio.NewReader(conn).Read(rdata)
  rdata = rdata[:n]
  if err_handle != nil {
    fmt.Println("/[read] ", err_handle)
  }

  split_rdata := bytes.Split(rdata, []byte("\r"))
  for _, data := range split_rdata {
    if len(data) >= 1 {
      payload := rcf_util.Topic_parse_client_read_payload(data, topic_name)
      split_payload := bytes.Split(payload, []byte("\nm"))
      for _, split_payload_msg := range split_payload {
        if len(split_payload_msg) >= 1 {
          msgs = append(msgs, split_payload_msg)
        }
      }
    }
  }
  return msgs
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func Topic_raw_data_subscribe(conn net.Conn, topic_name string) <-chan []byte{
  conn.Write([]byte(">topic-"+topic_name+"-subscribe-\r"))
  topic_listener := make(chan []byte)
  go func(topic_listener chan<- []byte){
    for {
      data := make([]byte, tcp_conn_buffer)
      n, _ := bufio.NewReader(conn).Read(data)
      data = data[:n]
      split_rdata := bytes.Split(data, []byte("\r"))

      for _, data := range split_rdata {
        if len(data)>=1 {
          topic_listener <- rcf_util.Topic_parse_client_read_payload(data, topic_name)
        }
      }
    }
  }(topic_listener)
  return topic_listener
}

// pushes string msg to topic msg stack
func Topic_publish_string_data(conn net.Conn, topic_name string, data string) {
  Topic_publish_raw_data(conn, topic_name, []byte(data))
}

// pulls x msgs from topic topic stack
func Topic_pull_string_data(conn net.Conn, nmsgs int, topic_name string) []string {
  var msgs []string
  payload_msgs := Topic_pull_raw_data(conn, nmsgs, topic_name)
  for _, payload_msg := range payload_msgs {
    if len(payload_msg) >= 1 {
      msgs = append(msgs, string(payload_msg))
    }
  }

  return msgs
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func Topic_string_data_subscribe(conn net.Conn, topic_name string) <-chan string{
  conn.Write([]byte(">topic-"+topic_name+"-subscribe-\r"))
  topic_listener := make(chan string)
  go func(topic_listener chan<- string){
    for {
      data := make([]byte, tcp_conn_buffer)
      n, _ := bufio.NewReader(conn).Read(data)
      data = data[:n]
      split_rdata := bytes.Split(data, []byte("\r"))

      for _, data := range split_rdata {
        if len(data)>=1 {
          topic_listener <- string(rcf_util.Topic_parse_client_read_payload(data, topic_name))
        }
      }
    }
  }(topic_listener)
  return topic_listener
}

// pushes data to topic stack
func Topic_publish_glob_data(conn net.Conn, topic_name string, data map[string]string) {
  encoded_data := []byte(rcf_util.Glob_map_encode(data).Bytes())
  Topic_publish_raw_data(conn, topic_name, encoded_data)
}

// pulls x msgs from topic topic stack
func Topic_pull_glob_data(conn net.Conn, nmsgs int, topic_name string) []map[string]string {
  glob_map := make([]map[string]string, 0)
  payload_msgs := Topic_pull_raw_data(conn, nmsgs, topic_name)
  for _, payload_msg := range payload_msgs {
    if len(payload_msg) >= 1 {
      glob_map = append(glob_map, rcf_util.Glob_map_decode(payload_msg))
    }
  }
  return glob_map
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func Topic_glob_data_subscribe(conn net.Conn, topic_name string) <-chan map[string]string{
  conn.Write([]byte(">topic-"+topic_name+"-subscribe-\r"))
  topic_listener := make(chan map[string]string)
  go func(topic_listener chan<- map[string]string ){
    for {
      data := make([]byte, tcp_conn_buffer)
      n, err := bufio.NewReader(conn).Read(data)
      data = data[:n]
      split_data := bytes.Split(data, []byte("\r"))
      for _,sdata := range split_data {
        if len(sdata) > 1 {
          payload := rcf_util.Topic_parse_client_read_payload(sdata, topic_name)

    		  data_map := rcf_util.Glob_map_decode(payload)
    		  topic_listener <- data_map
    		  if err != nil {
      			fmt.Println("conn closed")
      			break
    		  }
        }
      }
    }
  }(topic_listener)
  return topic_listener
}

//  creates new action on node
func Topic_create(conn net.Conn, topic_name string) {
  conn.Write([]byte(">topic-"+topic_name+"-create-\r"))
}

//  executes action
func Action_exec(conn net.Conn, action_name string, params []byte) {
  send_slice := append(append([]byte(">action-"+action_name+"-exec-"), params...), "\r"...)
  conn.Write(send_slice)
}

//  executes service
func Service_exec(conn net.Conn, service_name string, params []byte) []byte{
  conn.Write(append(append([]byte(">service-"+service_name+"-exec-"), params...), "\r"...))
  data := make([]byte, tcp_conn_buffer)
  for {
    n, err := bufio.NewReader(conn).Read(data)
    if err != nil {
      fmt.Println("service exec res rec err")
      break
    }
    if n != 0 {
      data = data[:n]
      break
    }
  }
  return rcf_util.Service_parse_client_read_payload(data, service_name)
}

// lists node's topics
func Topic_list(conn net.Conn) []string {
  conn.Write([]byte(">topic-all-list-\r"))
  data, _ := bufio.NewReader(conn).ReadString('\r')

  return strings.Split(data, ",")
}
