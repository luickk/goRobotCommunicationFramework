package rcf_cc

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

// pushes data to a topic
func push_data(data string, node_id int, topic_name string) {
  var conn net.Conn = connect_to_tcp_server(node_id)

  defer conn.Close()

  fmt.Fprintf(conn, topic_name+"+="+data + "\n")
}
