package rcf_cc_node_client

import(
  "fmt"
  "net"
  "strconv"
  "bufio"
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
func Push_data(conn net.Conn, data string, topic_name string) {
  conn.Write([]byte(topic_name+"+"+data + "\n"))
}

// pulls x elements from topic topic stack
func Pull_data(conn net.Conn, nelements int, topic_name string) []string {
  fmt.Fprintf(conn, topic_name+"-"+strconv.Itoa(nelements) + "\n")

  var elements []string

  for i:=0; i <= nelements; i++{
    rdata, _ := bufio.NewReader(conn).ReadString('\n')
    fmt.Println(rdata)
    elements = append(elements,rdata)
  }
  return elements
}

// lists node's topics
func List_cctopics(conn net.Conn) {
  conn.Write([]byte("list_cctopics\n"))
}

// pulls and pops x elements from topic topic stack
func Create_topic(conn net.Conn, topic_name string) {
  conn.Write([]byte("+"+topic_name + "\n"))
}
