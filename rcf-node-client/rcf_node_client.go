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

var tcpConnBuffer = 1024

// function to connect to tcp server (node) and returns connection
func connectToTcpServer(port int) net.Conn{
  conn, err := net.Dial("tcp4", ":"+strconv.Itoa(port))

  if err != nil {
    fmt.Println(err)
  }
  // don't forget to close connection
  return conn
}

// returns connection to node
func NodeOpenConn(node_id int) net.Conn {
  var conn net.Conn = connectToTcpServer(node_id)

  return conn
}

func NodeCloseConn(conn net.Conn) {
  conn.Write([]byte("end\r"))
  conn.Close()
}

// pushes raw byte slice msg to topic msg stack
func TopicPublishRawData(conn net.Conn, topicName string, data []byte) {
  send_slice := append(append([]byte(">topic-"+topicName+"-publish-"),data...),"\r"...)
  conn.Write(send_slice)
}

// pulls x msgs from topic topic stack
func TopicPullRawData(conn net.Conn, nmsgs int, topicName string) [][]byte {
  send_slice := append([]byte(">topic-"+topicName+"-pull-"+strconv.Itoa(nmsgs)), "\r"...)
  conn.Write(send_slice)
  var msgs [][]byte
  rdata := make([]byte, tcpConnBuffer)
  n, err_handle := bufio.NewReader(conn).Read(rdata)
  rdata = rdata[:n]
  if err_handle != nil {
    fmt.Println("/[read] ", err_handle)
  }

  splitRData := bytes.Split(rdata, []byte("\r"))
  for _, data := range splitRData {
    if len(data) >= 1 {
      payload := rcf_util.TopicParseClientReadPayload(data, topicName)
      split_payload := bytes.Split(payload, []byte("\nm"))
      for _, splitPayloadMsg := range split_payload {
        if len(splitPayloadMsg) >= 1 {
          msgs = append(msgs, splitPayloadMsg)
        }
      }
    }
  }
  return msgs
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicRawDataSubscribe(conn net.Conn, topicName string) <-chan []byte{
  conn.Write([]byte(">topic-"+topicName+"-subscribe-\r"))
  topicListener := make(chan []byte)
  go func(topicListener chan<- []byte){
    for {
      data := make([]byte, tcpConnBuffer)
      n, _ := bufio.NewReader(conn).Read(data)
      data = data[:n]
      splitRData := bytes.Split(data, []byte("\r"))

      for _, data := range splitRData {
        if len(data)>=1 {
          topicListener <- rcf_util.TopicParseClientReadPayload(data, topicName)
        }
      }
    }
  }(topicListener)
  return topicListener
}

// pushes string msg to topic msg stack
func TopicPublishStringData(conn net.Conn, topicName string, data string) {
  TopicPublishRawData(conn, topicName, []byte(data))
}

// pulls x msgs from topic topic stack
func TopicPullStringData(conn net.Conn, nmsgs int, topicName string) []string {
  var msgs []string
  payloadMsgs := TopicPullRawData(conn, nmsgs, topicName)
  for _, payloadMsg := range payloadMsgs {
    if len(payloadMsg) >= 1 {
      msgs = append(msgs, string(payloadMsg))
    }
  }

  return msgs
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicStringDataSubscribe(conn net.Conn, topicName string) <-chan string{
  conn.Write([]byte(">topic-"+topicName+"-subscribe-\r"))
  topicListener := make(chan string)
  go func(topicListener chan<- string){
    for {
      data := make([]byte, tcpConnBuffer)
      n, _ := bufio.NewReader(conn).Read(data)
      data = data[:n]
      splitRData := bytes.Split(data, []byte("\r"))

      for _, data := range splitRData {
        if len(data)>=1 {
          topicListener <- string(rcf_util.TopicParseClientReadPayload(data, topicName))
        }
      }
    }
  }(topicListener)
  return topicListener
}

// pushes data to topic stack
func TopicPublishGlobData(conn net.Conn, topicName string, data map[string]string) {
  encodedData := []byte(rcf_util.GlobMapEncode(data).Bytes())
  // println("Published data:")
  // fmt.Printf("%08b", encodedData)
  TopicPublishRawData(conn, topicName, encodedData)
}

// pulls x msgs from topic topic stack
func TopicPullGlobData(conn net.Conn, nmsgs int, topicName string) []map[string]string {
  glob_map := make([]map[string]string, 0)
  payloadMsgs := TopicPullRawData(conn, nmsgs, topicName)
  for _, payloadMsg := range payloadMsgs {
    if len(payloadMsg) >= 1 {
      glob_map = append(glob_map, rcf_util.GlobMapDecode(payloadMsg))
    }
  }
  return glob_map
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicGlobDataSubscribe(conn net.Conn, topicName string) <-chan map[string]string{
  conn.Write([]byte(">topic-"+topicName+"-subscribe-\r"))
  topicListener := make(chan map[string]string)
  go func(topicListener chan<- map[string]string ){
    for {
      data := make([]byte, tcpConnBuffer)
      n, err := bufio.NewReader(conn).Read(data)
      data = data[:n]
      split_data := bytes.Split(data, []byte("\r"))
      for _,sdata := range split_data {
        if len(sdata) > 1 {
          payload := rcf_util.TopicParseClientReadPayload(sdata, topicName)
          if len(payload) > 1 {
            data_map := rcf_util.GlobMapDecode(payload)
            topicListener <- data_map
            if err != nil {
              fmt.Println("conn closed")
              break
            }
          }
        }
      }
    }
  }(topicListener)
  return topicListener
}

//  creates new action on node
func TopicCreate(conn net.Conn, topicName string) {
  conn.Write([]byte(">topic-"+topicName+"-create-\r"))
}

//  executes action
func ActionExec(conn net.Conn, actionName string, params []byte) {
  send_slice := append(append([]byte(">action-"+actionName+"-exec-"), params...), "\r"...)
  conn.Write(send_slice)
}

//  executes service
func ServiceExec(conn net.Conn, serviceName string, params []byte) []byte {
  data := make([]byte, tcpConnBuffer)
  var payload []byte
  executed := true
  for {
    n, err := bufio.NewReader(conn).Read(data)
    if err != nil {
      fmt.Println("service exec res rec err")
      break
    }
    if executed { 
      conn.Write(append(append([]byte(">service-"+serviceName+"-exec-"), params...), "\r"...))
      executed = false
    }
    if n != 0 {
      payload = rcf_util.ServiceParseClientReadPayload(data[:n], serviceName)
      if len(payload) >= 1 {
        break
      }
    }
  }
  return payload
}

// lists node's topics
func TopicList(conn net.Conn) []string {
  conn.Write([]byte(">topic-all-list-\r"))
  data, _ := bufio.NewReader(conn).ReadString('\r')

  return strings.Split(data, ",")
}
