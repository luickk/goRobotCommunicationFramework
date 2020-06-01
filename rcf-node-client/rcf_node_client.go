package rcf_node_client

import(
	"fmt"
	"net"
	"strconv"
	"strings"
	"bufio"
	"bytes"
  "math/rand"
	"rcf/rcf-util"
)

var tcpConnBuffer = 1024
var connChannel chan []byte

// parses information and pushes them to byte channel 
func connHandler(conn net.Conn, connChan chan []byte) {
  for {
    data := make([]byte, tcpConnBuffer)
    n, _ := bufio.NewReader(conn).Read(data)
    data = data[:n]
    splitRData := bytes.Split(data, []byte("\r"))
    for _, data := range splitRData {
      if len(data)>=1 {
        connChan <- data
      }
    }
  }
}
// function to connect to tcp server (node)
// returns connHandler channel, to which incoming parsed data is pushed 
func connectToTcpServer(port int) (chan []byte, net.Conn) {
  conn, err := net.Dial("tcp4", ":"+strconv.Itoa(port))
  connChannel = make(chan []byte)
  go connHandler(conn, connChannel)
  
  if err != nil {
    fmt.Println(err)
  }
  // don't forget to close connection
  return connChannel, conn
}

// returns connection to node
func NodeOpenConn(node_id int) (chan []byte, net.Conn) {
  return connectToTcpServer(node_id)
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
func TopicPullRawData(conn net.Conn, connChannel chan []byte, nmsgs int, topicName string) [][]byte {
  pullReqId := rand.Intn(255)
  if pullReqId == 0 || pullReqId == 2 {
    pullReqId = rand.Intn(255)  
  }
  send_slice := append([]byte(">topic-"+topicName+","+strconv.Itoa(pullReqId)+"-pull-"+strconv.Itoa(nmsgs)), "\r"...)
  conn.Write(send_slice)
  var msgs [][]byte
  receivedReply := false
  for !receivedReply {
    select {
      case data := <-connChannel:
        if len(data) >= 1 {
          payload := rcf_util.TopicParseSinglePullClientReadPayload(data, pullReqId, topicName)
          split_payload := bytes.Split(payload, []byte("\nm"))
          for _, splitPayloadMsg := range split_payload {
            if len(splitPayloadMsg) >= 1 {
              msgs = append(msgs, splitPayloadMsg)
            }
          }
          receivedReply = true
          break
        }
    }
  }
  return msgs
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicRawDataSubscribe(conn net.Conn, connChannel chan []byte, topicName string) <-chan []byte{
  conn.Write([]byte(">topic-"+topicName+"-subscribe-\r"))
  topicListener := make(chan []byte)
  go func(topicListener chan<- []byte){
    for {
      select {
        case data := <-connChannel:
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
func TopicPullStringData(conn net.Conn, connChannel chan []byte, nmsgs int, topicName string) []string {
  var msgs []string
  payloadMsgs := TopicPullRawData(conn, connChannel, nmsgs, topicName)
  for _, payloadMsg := range payloadMsgs {
    if len(payloadMsg) >= 1 {
      msgs = append(msgs, string(payloadMsg))
    }
  }

  return msgs
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicStringDataSubscribe(conn net.Conn, connChannel chan []byte, topicName string) <-chan string{
  conn.Write([]byte(">topic-"+topicName+"-subscribe-\r"))
  topicListener := make(chan string)
  go func(topicListener chan<- string){
    for {
      select {
        case data := <-connChannel:
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
func TopicPullGlobData(conn net.Conn, connChannel chan []byte, nmsgs int, topicName string) []map[string]string {
  glob_map := make([]map[string]string, 0)
  payloadMsgs := TopicPullRawData(conn, connChannel, nmsgs, topicName)
  for _, payloadMsg := range payloadMsgs {
    if len(payloadMsg) >= 1 {
      glob_map = append(glob_map, rcf_util.GlobMapDecode(payloadMsg))
    }
  }
  return glob_map
}

// waits continuously for incoming topic msgs, enables topic data streaming before
func TopicGlobDataSubscribe(conn net.Conn, connChannel chan []byte, topicName string) <-chan map[string]string{
  conn.Write([]byte(">topic-"+topicName+"-subscribe-\r"))
  topicListener := make(chan map[string]string)
  go func(topicListener chan<- map[string]string ){
    for {
      select {
        case sdata := <-connChannel:
          if len(sdata) >= 1 {
            payload := rcf_util.TopicParseClientReadPayload(sdata, topicName)
            if len(payload) >= 1 {
              data_map := rcf_util.GlobMapDecode(payload)
              topicListener <- data_map
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

// executes service and returns channel to which the results are pushed
// each service has an assigned id to prohibit result collisions
func ServiceExec(conn net.Conn, connChannel chan []byte, serviceName string, params []byte) <-chan []byte {
  serviceListener := make(chan []byte)
  go func(chan []byte) {
    serviceId := rand.Intn(255)
    if serviceId == 0 || serviceId == 2 {
      serviceId = rand.Intn(255)  
    }
    var payload []byte
    requested := false
    foundReply := false
    for !foundReply{
      select {
        case data := <-connChannel:
          if len(data) >= 1 {
            if !requested {
              conn.Write(append(append([]byte(">service-"+serviceName+","+strconv.Itoa(serviceId)+"-exec-"), params...), "\r"...))
              requested = true
            }
            payload = rcf_util.ServiceParseClientReadPayload(data, serviceName, serviceId)
            if len(payload) >= 1 {
              serviceListener <- payload
              foundReply = true
              break
            }  
          }
      }
    }
  }(serviceListener)
  return serviceListener
}

// lists node's topics
func TopicList(conn net.Conn, connChannel chan []byte) []string {
  conn.Write([]byte(">topic-all-list-\r"))
  data := <-connChannel
  
  return strings.Split(string(data), ",")
}
