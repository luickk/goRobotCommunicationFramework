package rcf_cc_node_client

import(
  "fmt"
  "net"
  "strconv"
  "strings"
  "bufio"
  "robot-communication-framework/rcf_util"
)

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

// returns conncetion to node
func Connect_to_cc_node(node_id int) net.Conn {
  var conn net.Conn = connect_to_tcp_server(node_id)

  return conn
}

// pushes data to topic topic stack
func Push_data(conn net.Conn, topic_name string, data string) {
  conn.Write([]byte(topic_name+"+"+data+"\n"))
}

// pulls x elements from topic topic stack
func Pull_data(conn net.Conn, nelements int, topic_name string) []string {
  conn.Write([]byte(topic_name+"-"+strconv.Itoa(nelements) + "\n"))

  var elements []string

  rdata, _ := bufio.NewReader(conn).ReadString('\n')

  elements = strings.Split(rcf_util.Trim_suffix(rdata), ",")

  return elements
}

// lists node's topics
func List_cctopics(conn net.Conn) []string {
  conn.Write([]byte("list_cctopics\n"))
  data, _ := bufio.NewReader(conn).ReadString('\n')

  return strings.Split(rcf_util.Trim_suffix(data), ",")
}

// waits continuously for incoming topic elements, enables topic data streaming before
func Continuous_data_pull(conn net.Conn, topic_name string) <-chan string{
  conn.Write([]byte("$"+topic_name+"\n"))
  topic_listener := make(chan string)
  go func(topic_listener chan<- string) chan<- string{
    for {
      data, _ := bufio.NewReader(conn).ReadString('\n')
      topic_listener <- data
      fmt.Println("data changed")
    }
  }(topic_listener)
  return topic_listener
}

//  creates new topic on node
func Create_topic(conn net.Conn, topic_name string) {
  conn.Write([]byte("+"+topic_name + "\n"))
}

func Close_cc_node(conn net.Conn) {
  conn.Write([]byte("end\n"))
  conn.Close()
}
