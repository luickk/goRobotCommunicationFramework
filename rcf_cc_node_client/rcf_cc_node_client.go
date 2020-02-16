package rcf_cc_node_client

import(
  "fmt"
  "net"
  "strconv"
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

// pushes data to topic topic stack
func Push_data(data string, node_id int, topic_name string) {
  var conn net.Conn = connect_to_tcp_server(node_id)

  defer fmt.Fprintf(conn, "end"+"\n")
  defer conn.Close()

  fmt.Fprintf(conn, topic_name+"+"+data + "\n")
}

// pulls and pops x elements from topic topic stack
func Pop_data(elements int, node_id int, topic_name string) {
  var conn net.Conn = connect_to_tcp_server(node_id)

  defer fmt.Fprintf(conn, "end"+"\n")
  defer conn.Close()

  fmt.Fprintf(conn, topic_name+"-"+strconv.Itoa(elements) + "\n")


}

// pulls and pops x elements from topic topic stack
func Create_topic(topic_name string, node_id int) {
  var conn net.Conn = connect_to_tcp_server(node_id)

  defer fmt.Fprintf(conn, "end"+"\n")
  defer conn.Close()

  fmt.Fprintf(conn, "+"+topic_name + "\n")


}
