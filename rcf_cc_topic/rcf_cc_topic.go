package rcf_cc_topic

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

// pushes data to topic queue
func Push_data(data string, node_id int, topic_name string) {
  var conn net.Conn = connect_to_tcp_server(node_id)

  defer fmt.Fprintf(conn, "end"+"\n")
  defer conn.Close()

  fmt.Fprintf(conn, topic_name+"+"+data + "\n")
}

// pulls and pops x elements from topic queue
func Pop_data(data string, elements int, node_id int, topic_name string) {
  var conn net.Conn = connect_to_tcp_server(node_id)

  defer fmt.Fprintf(conn, "end"+"\n")
  defer conn.Close()

  fmt.Fprintf(conn, topic_name+"-"+strconv.Itoa(elements) + "\n")
  for i:=0; i<=elements; i++ {
      message, _ := bufio.NewReader(conn).ReadString('\n')
      fmt.Println("received element ", strconv.Itoa(elements), " from topic: ", topic_name)
      fmt.Printf(message)
  }

}
