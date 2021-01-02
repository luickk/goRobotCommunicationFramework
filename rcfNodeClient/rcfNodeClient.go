/*
Package rcfnodeclient implements all functions to communicate  with a rcf node.
*/
package rcfnodeclient

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"rcf/rcfUtil"
)

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 1024

// requestCapacity defines the maximum capacity for active service, topic requests in queue
var requestCapacity = 1000

// struct that contains all information for a valid request for the handler which communicates with the give node
type dataRequest struct {
	// name always includes optional id
	Name string
	// or action
	Op string
	// true if request has been processed
	Fulfilled bool
	// req id
	Id int
	// the result of the request
	ReturnedPayload chan []byte
	// a pull operation requires a 2 dim slice, since it can contain multiple msgs
	PullOpReturnedPayload chan [][]byte
}

// client struct contains the connection to the node for write access and the request channels which are read/ processed by the handlers
type Client struct {
	// connection to the node
	Conn net.Conn

	// clientWriteRequestCh is the channel alle write requests are written to
	clientWriteRequestCh chan []byte

	// topic pull/ sub request channel which is read/ processed by the topic handler
	TopicContextRequests chan dataRequest
	// service call request channel which is read/ processed by the serice handler
	ServiceContextRequests chan dataRequest
}

// topicContextMsgs is the channel wich raw msgs from the node are pushed to if their type/ context is topic
var topicContextMsgs chan []rcfUtil.Smsg

// serviceContextMsgs is the channel wich raw msgs from the node are pushed to if their type/ context is service
var serviceContextMsgs chan []rcfUtil.Smsg

// basic logger declarations
var (
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
)

// parses incoming instructions from the node and sorts them according to their context/ type
// pushes sorted instructions to the according handler
func (client *Client)connHandler(conn net.Conn, topicContextMsgs chan rcfUtil.Smsg, serviceContextMsgs chan rcfUtil.Smsg) {
	InfoLogger.Println("handleConnection started")
	defer conn.Close()
	netDataBuffer := make([]byte, tcpConnBuffer)
	decodedMsg := new(rcfUtil.Smsg)
	for {
		netDataBuffer, err := bufio.NewReader(conn).ReadBytes("0x0")
		if err != nil {
			ErrorLogger.Println("handleConnection routine reading error")
			ErrorLogger.Println(err)
			break
		}
		// parsing instrucitons from client
		if err := rcfUtil.DecodeMsg(decodedMsg, netDataBuffer); err != nil {
			WarningLogger.Println(err)
		}
		switch decodedMsg.Type {
		case "topic":
			topicContextMsgs <- decodedMsg
 		case "service":
			serviceContextMsgs <- decodedMsg
		}
	}
}

// clientWriteRequestHandler handles all write request to clients
func (client *Client)clientWriteRequestHandler() {
	InfoLogger.Println("writeHandler started")
	for {
		select {
		case writeRequest := <-client.clientWriteRequestCh:
			client.Conn.Write(append(writeRequest, []byte{0x00}))
		}
	}
}

// handles topic pull/ sub requests and processes topic context/ type msg payloads
func (client *Client)topicHandler(topicContextMsgs chan []rcfUtil.Smsg, topicRequests chan dataRequest) {
	InfoLogger.Println("topicHandler started")
	connClosed := false
	requests := make(map[string]dataRequest, 1000)
	for {
		select {
		case decodedMsg := <-topicContextMsgs:
			for name, req := range requests {
				if operation == "pull" {
					if req.Op == "pull" && req.Fulfilled == false {
						if requestFound {
							req.PullOpReturnedPayload <- decodedMsg.Payload
							req.Fulfilled = true
							requests[name] = req
							delete(requests, name)
						}
				} else if operation == "sub" {
					if req.Op == "sub" && req.Fulfilled == false {
						var payload []byte
						if decodedMsg.Name == name {
							req.ReturnedPayload <- decodedMsg.payload
						}
					}
				} else if operation == "pullinfo" {
					if req.Op == "pulltopiclist" && req.Fulfilled == false {
						var payload []byte
						req.ReturnedPayload <- []byte("err")
						WarningLogger.Println("topicHandler info topicList parsing payload extraction error")
						req.Fulfilled = true
						delete(requests, name)
					}
				}
			}
		case request := <-topicRequests:
			if connClosed {
				if request.Op == "pulltopiclist" || request.Op == "sub" {
					request.ReturnedPayload <- []byte("conn closed")
				} else if request.Op == "pull" {
					request.PullOpReturnedPayload <- [][]byte{[]byte("conn closed")}
				}
				break
			}
			requests[request.Name] = request
		}
	}
}

// handles service call requests and processes the results which are contained in the service type/context msg payloads
func (client *Client)serviceHandler(serviceContextMsgs <-chan rcfUtil.Smsg, serviceRequests <-chan dataRequest) {
	connClosed := false
	requests := make(map[string]dataRequest, 1000)
	for {
		select {
		case decodedMsg := <-serviceContextMsgs:
			for name, req := range requests {
				if req.Fulfilled == false {
					req.ReturnedPayload <- payload
					req.Fulfilled = true
					requests[name] = req
					delete(requests, name)
				}
			}
		case request := <-serviceRequests:
			requests[request.Name] = request
		}
	}
}

// ServiceExec executes service and returns channel to which the results are pushed
// each service has an assigned id to prohibit result collisions
func (client *Client)ServiceExec(serviceName string, params []byte) []byte {
	serviceID := rcfUtil.GenRandomIntID()

	encodingMsg := new(rcfUtil.Smsg)

	request := new(dataRequest)
	request.Name = serviceName
	request.Id = serviceID
	request.Op = "exec"
	request.Fulfilled = false
	request.ReturnedPayload = make(chan []byte)
	client.ServiceContextRequests <- *request


	encodingMsg.Type = "service"
	encodingMsg.Name = serviceName
	encodingMsg.Id = serviceID
	encodingMsg.Operation = "exec"
	encodingMsg.payload = params
	client.clientWriteRequestCh <- rcfUtil.EncodeMsg(encodingMsg)


	reply := false
	payload := []byte{}

	for !reply {
		select {
		case liveDataRes := <-request.ReturnedPayload:
			payload = liveDataRes
			reply = true
			break
		}
	}
	return payload
}

// TopicPullRawData Pulls raw data msgs from given topic
func (client *Client)TopicPullRawData(topicName string, nmsgs int) [][]byte {
	// generates random id for the name
	pullReqID := rcfUtil.GenRandomIntID()

	// creating request for the payload which is sent back from the node
	request := new(dataRequest)
	request.Name = topicName
	request.Op = "pull"
	request.Id = pullReqID
	request.Fulfilled = false
	request.PullOpReturnedPayload = make(chan [][]byte)
	// pushing request to topic handler where it is process
	client.TopicContextRequests <- *request

	// create instrucitons slice for the node according to the protocl
	encodingMsg.Type = "topic"
	encodingMsg.Name = topicName
	encodingMsg.Id = pullReqID
	encodingMsg.Operation = "pull"
	encodingMsg.payload = []byte{}
	client.clientWriteRequestCh <- rcfUtil.EncodeMsg(encodingMsg)

	reply := false
	payload := [][]byte{}

	// wainting for request to be processed and retrieval of payload
	for !reply {
		select {
		case liveDataRes := <-request.PullOpReturnedPayload:
			payload = liveDataRes
			reply = true
			close(request.PullOpReturnedPayload)
			break
		}
	}
	return payload
}

// TopicRawDataSubscribe subscribes to topic and pulls raw msgs data
func (client *Client)TopicRawDataSubscribe(topicName string) chan []byte {
	// generating random id for the name
	pullReqID := rcfUtil.GenRandomIntID()

	// creating request for topic handler
	request := new(dataRequest)
	request.Name = topicName
	request.Id = pullReqID
	request.Op = "sub"
	request.Fulfilled = false
	request.ReturnedPayload = make(chan []byte)
	// sending request to topic handler
	client.TopicContextRequests <- *request

	// creating and writing instruction slice for the node
	encodingMsg.Type = "topic"
	encodingMsg.Name = topicName
	encodingMsg.Id = pullReqID
	encodingMsg.Operation = "subscribe"
	encodingMsg.payload = []byte{}
	client.clientWriteRequestCh <- rcfUtil.EncodeMsg(encodingMsg)

	// returning channel from request to which the topic handler writes the results
	return request.ReturnedPayload
}

// connectToTCPServer function to connect to tcp server (node)
// returns connHandler channel, to which incoming parsed data is pushed
func (client *Client)connectToTCPServer(port int) (net.Conn, chan int, chan []byte, chan []byte, error) {
	topicContextMsgs = make(chan []rcfUtil.Smsg)
	serviceContextMsgs = make(chan []rcfUtil.Smsg)
	conn, err := net.Dial("tcp4", ":"+strconv.Itoa(port))

	if err != nil {
		ErrorLogger.Println("connectToTcpServer could not connect to tcp server (node instance)")
		ErrorLogger.Println(err)
	} else {
		go client.connHandler(conn, topicContextMsgs, serviceContextMsgs)
	}

	// don't forget to close connection
	return conn, topicContextMsgs, serviceContextMsgs, err
}

// NodeOpenConn initiates loggers and comm channels for handler and start handlers
// returns client struct which defines relevant information for the interface functions to work
func New(nodeID int) (Client, error) {
	InfoLogger = log.New(os.Stdout, "[CLIENT] INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "[CLIENT] WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "[CLIENT] ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger.SetOutput(ioutil.Discard)

	rcfUtil.InfoLogger = InfoLogger
	rcfUtil.WarningLogger = WarningLogger
	rcfUtil.ErrorLogger = ErrorLogger

	conn, topicContextMsgs, serviceContextMsgs, err := connectToTCPServer(nodeID)
	if err != nil {
		connected = false
	}
	topicContextRequests := make(chan dataRequest)
	serviceContextRequests := make(chan dataRequest)

	go client.topicHandler(topicContextMsgs, topicContextRequests)
	go client.serviceHandler(serviceContextMsgs, serviceContextRequests)

	client := new(Client)
	client.Conn = conn
	client.clientWriteRequestCh = make(chan []byte, 100)
	client.TopicContextRequests = topicContextRequests
	client.ServiceContextRequests = serviceContextRequests

	go client.clientWriteRequestHandler()
	return *client, err
}

// TopicPublishRawData pushes raw byte slice msg to topic msg stack
func (client *Client)TopicPublishRawData(topicName string, data []byte) {

	encodingMsg.Type = "topic"
	encodingMsg.Name = topicName
	encodingMsg.Operation = "publish"
	encodingMsg.payload = data
	client.clientWriteRequestCh <- rcfUtil.EncodeMsg(encodingMsg)
}

// ActionExec executes action
func (client *Client)ActionExec(actionName string, params []byte) {
	encodingMsg := new(rcfUtil.Smsg)
	encodingMsg.Type = "action"
	encodingMsg.Name = actionName
	encodingMsg.Operation = "exec"
	encodingMsg.payload = []byte{}
	client.clientWriteRequestCh <- rcfUtil.EncodeMsg(encodingMsg)
}

// TopicCreate creates new action on node
func (client *Client)TopicCreate(topicName string) {
	encodingMsg := new(rcfUtil.Smsg)
	encodingMsg.Type = "topic"
	encodingMsg.Name = topicName
	encodingMsg.Operation = "create"
	encodingMsg.payload = []byte{}
	client.clientWriteRequestCh <- rcfUtil.EncodeMsg(encodingMsg)
}

// TopicList lists node's topics
func (client *Client)TopicList() []string {

	encodingMsg := new(rcfUtil.Smsg)
	encodingMsg.Type = "topic"
	encodingMsg.Name = "all"
	encodingMsg.Operation = "list"
	encodingMsg.payload = []byte{}

	client.clientWriteRequestCh <- rcfUtil.EncodeMsg(encodingMsg)

	// creating request for the payload which is sent back from the node
	request := new(dataRequest)
	request.Name = "topiclist"
	request.Op = "pulltopiclist"
	request.Fulfilled = false
	request.ReturnedPayload = make(chan []byte)
	// pushing request to topic handler where it is process
	client.TopicContextRequests <- *request

	reply := false
	payload := []byte{}

	// wainting for request to be processed and retrieval of payload
	for !reply {
		select {
		case liveDataRes := <-request.ReturnedPayload:
			payload = liveDataRes
			reply = true
			break
		}
	}
	return strings.Split(string(payload), ",")
}
