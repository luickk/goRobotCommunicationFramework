package rcf_node_client

import(
  "fmt"
  "net"
  "strconv"
  "strings"
  "bufio"
  "bytes"
  "robot-communication-framework/rcf_util"
)

var tcp_conn_buffer = 1024

// function to connect to tcp server (node) and returns connection
func connect_to_tcp_server(port int) net.Conn{
  conn, err := net.Dial("tcp4", ":"+strconv.Itoa(port))

  if err != nil {
    fmt.Println("an error occured: ")
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

// pushes data to topic stack
func Topic_publish_data(conn net.Conn, topic_name string, data string) {
  conn.Write([]byte(topic_name+"+"+data+"\r"))
}

// pulls x elements from topic topic stack
func Topic_pull_data(conn net.Conn, nelements int, topic_name string) []string {
  conn.Write([]byte(topic_name+"-"+strconv.Itoa(nelements) + "\r"))
  var elements []string
  rdata := make([]byte, tcp_conn_buffer)
  n, err_handle := bufio.NewReader(conn).Read(rdata)
  rdata = rdata[:n]

  if err_handle != nil {
    fmt.Println("/[read] ", err_handle)
  }

  split_rdata := bytes.Split(rdata, []byte("\r"))
  for _, data := range split_rdata {
    payload := rcf_util.Topic_parse_client_read_protocol(data, topic_name)
  	split_payload := bytes.Split(payload, []byte("\nm"))
    for _, split_payload_msg := range split_payload {
      if len(split_payload_msg) >= 1 {
        elements = append(elements, string(split_payload_msg))
      }
    }
  }

  return elements
}

// pushes data to topic stack
func Topic_glob_publish_data(conn net.Conn, topic_name string, data map[string]string) {
  encoded_data := []byte(rcf_util.Glob_map_encode(data).Bytes())
  bsend := append([]byte(topic_name+"+"), encoded_data...)
  bsend = append(bsend, []byte("\r")...)
  conn.Write(bsend)
}

// pulls x elements from topic topic stack
func Topic_glob_pull_data(conn net.Conn, nelements int, topic_name string) []map[string]string {
  conn.Write([]byte(topic_name+"-"+strconv.Itoa(nelements) + "\r"))
  elements := make([]map[string]string, 0)
  rdata := make([]byte, tcp_conn_buffer)
  n, err_handle := bufio.NewReader(conn).Read(rdata)
  rdata = rdata[:n]
  if err_handle != nil {
    fmt.Println("/[read] ", err_handle)
  }

  split_rdata := bytes.Split(rdata, []byte("\r"))

  for _, data := range split_rdata {
    payload := rcf_util.Topic_parse_client_read_protocol(data, topic_name)
  	split_payload := bytes.Split(payload, []byte("\nm"))
    for _, split_payload_msg := range split_payload {
      if len(split_payload_msg) > 1 {
        elements = append(elements, rcf_util.Glob_map_decode(split_payload_msg))
      }
    }
  }
  return elements
}

// waits continuously for incoming topic elements, enables topic data streaming before
func Topic_subscribe(conn net.Conn, topic_name string) <-chan string{
  conn.Write([]byte("$"+topic_name+"\r"))
  topic_listener := make(chan string)
  go func(topic_listener chan<- string){
    for {
      data := make([]byte, tcp_conn_buffer)
      n, _ := bufio.NewReader(conn).Read(data)
      data = data[:n]
      split_rdata := bytes.Split(data, []byte("\r"))

      for _, data := range split_rdata {
        if len(data)>=1 {
          topic_listener <- string(rcf_util.Topic_parse_client_read_protocol(data, topic_name))
        }
      }
    }
  }(topic_listener)
  return topic_listener
}
// waits continuously for incoming topic elements, enables topic data streaming before
func Topic_glob_subscribe(conn net.Conn, topic_name string) <-chan map[string]string{
  conn.Write([]byte("$"+topic_name+"\r"))
  topic_listener := make(chan map[string]string)
  go func(topic_listener chan<- map[string]string ){
    for {
      data := make([]byte, tcp_conn_buffer)
      n, err := bufio.NewReader(conn).Read(data)
      data = data[:n]
      split_data := bytes.Split(data, []byte("\r"))
      for _,sdata := range split_data {
        if len(sdata) > 1 {
          payload := rcf_util.Topic_parse_client_read_protocol(sdata, topic_name)

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
  conn.Write([]byte("+"+topic_name + "\r"))
}

//  executes action
func Action_exec(conn net.Conn, action_name string) {
  conn.Write([]byte("*"+action_name + "\r"))
}

//  executes service
func Service_exec(conn net.Conn, action_name string) []byte{
  conn.Write([]byte("#"+action_name + "\r"))
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
  return data
}

// lists node's topics
func Topic_list(conn net.Conn) []string {
  conn.Write([]byte("list_topics\r"))
  data, _ := bufio.NewReader(conn).ReadString('\r')

  return strings.Split(data, ",")
}
