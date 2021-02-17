/*
Package rcfnodeclient implements all functions to communicate  with a rcf node.
*/
package rcfNodeClient

import (
	"bufio"
	"net"
	"strconv"

	"goRobotCommunicationFramework/rcfUtil"
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
var topicContextMsgs chan rcfUtil.Smsg

// serviceContextMsgs is the channel wich raw msgs from the node are pushed to if their type/ context is service
var serviceContextMsgs chan rcfUtil.Smsg

// parses incoming instructions from the node and sorts them according to their context/ type
// pushes sorted instructions to the according handler
func (client *Client)connHandler(conn net.Conn, topicContextMsgs chan rcfUtil.Smsg, serviceContextMsgs chan rcfUtil.Smsg, errorStream chan error) {
	defer conn.Close()
	// routineId := rcfUtil.GenRandomIntID()
	netDataBuffer := make([]byte, tcpConnBuffer)
	decodedMsg := new(rcfUtil.Smsg)
	var err error
	for {
		netDataBuffer, err = rcfUtil.ReadFrame(conn)
		if err != nil {
			errorStream <- err
			return
		}
		// parsing instrucitons from client
		if err = rcfUtil.DecodeMsg(decodedMsg, netDataBuffer); err != nil {
			errorStream <- err
			return
		}
		switch decodedMsg.Type {
		case "topic":
			topicContextMsgs <- *decodedMsg
 		case "service":
			serviceContextMsgs <- *decodedMsg
		}
		netDataBuffer = []byte{}
	}
	decodedMsg = nil
	return
}

// clientWriteRequestHandler handles all write request to clients
func (client *Client)clientWriteRequestHandler(errorStream chan error) {
	writer := bufio.NewWriter(client.Conn)
	for {
		select {
		case writeRequest := <-client.clientWriteRequestCh:
			if err := rcfUtil.WriteFrame(writer, writeRequest); err != nil {
				errorStream <- err
				return
			}
		}
	}
	return
}

// handles topic pull/ sub requests and processes topic context/ type msg payloads
func (client *Client)topicHandler(topicContextMsgs chan rcfUtil.Smsg, topicRequests chan dataRequest) {
	requests := make(map[int]dataRequest)
	for {
		select {
		case decodedMsg := <-topicContextMsgs:
			for id, req := range requests {
				switch decodedMsg.Operation {
				case "pull":
					if req.Op == "pull" && req.Fulfilled == false {
						if req.Id == decodedMsg.Id {
							req.PullOpReturnedPayload <- decodedMsg.MultiplePayload
							req.Fulfilled = true
							requests[id] = req
							delete(requests, id)
						}
					}
				case "sub":
					if req.Op == "sub" && req.Fulfilled == false {
						if decodedMsg.Name == req.Name {
							req.ReturnedPayload <- decodedMsg.Payload
						}
					}
				case "pullinfo":
					if req.Op == "pulltopiclist" && req.Fulfilled == false {
            req.PullOpReturnedPayload <- decodedMsg.MultiplePayload
						req.Fulfilled = true
						delete(requests, id)
					}
				}
			}
		case req := <-topicRequests:
			requests[req.Id] = req
		}
	}
}

// handles service call requests and processes the results which are contained in the service type/context msg payloads
func (client *Client)serviceHandler(serviceContextMsgs <-chan rcfUtil.Smsg, serviceRequests <-chan dataRequest) {
	requestsById := make(map[int]dataRequest)
	for {
		select {
		case decodedMsg := <-serviceContextMsgs:
			if req, ok := requestsById[decodedMsg.Id]; ok && !req.Fulfilled {
				req.ReturnedPayload <- decodedMsg.Payload
				req.Fulfilled = true
				requestsById[req.Id] = req
				delete(requestsById, req.Id)
			}
		case request := <-serviceRequests:
			requestsById[request.Id] = request
		}
	}
}

// ServiceExec executes service and returns channel to which the results are pushed
// each service has an assigned id to prohibit result collisions
func (client *Client)ServiceExec(serviceName string, params []byte) ([]byte, error) {
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
	encodingMsg.Payload = params
	encodedMsg, err := rcfUtil.EncodeMsg(encodingMsg)
	if err != nil {
		return []byte{}, err
	}
	client.clientWriteRequestCh <- encodedMsg


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
	request = nil
	return payload, nil
}

// TopicPullRawData Pulls raw data msgs from given topic
func (client *Client)TopicPullData(topicName string, nmsgs int) ([][]byte, error) {
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
	encodingMsg := new(rcfUtil.Smsg)
	encodingMsg.Type = "topic"
	encodingMsg.Name = topicName
	encodingMsg.Id = pullReqID
	encodingMsg.Operation = "pull"
	encodingMsg.Payload = []byte(strconv.Itoa(nmsgs))
	encodedMsg, err := rcfUtil.EncodeMsg(encodingMsg)
	if err != nil {
		return [][]byte{}, err
	}
	client.clientWriteRequestCh <- encodedMsg
	encodingMsg = nil
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
	request = nil
	return payload, nil
}

// TopicRawDataSubscribe subscribes to topic and pulls raw msgs data
func (client *Client)TopicDataSubscribe(topicName string) (chan []byte, error) {
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
	encodingMsg := new(rcfUtil.Smsg)
	encodingMsg.Type = "topic"
	encodingMsg.Name = topicName
	encodingMsg.Id = pullReqID
	encodingMsg.Operation = "subscribe"
	encodingMsg.Payload = []byte{}
	encodedMsg, err := rcfUtil.EncodeMsg(encodingMsg)
	if err != nil {
		return request.ReturnedPayload, err
	}
	client.clientWriteRequestCh <- encodedMsg
	encodingMsg = nil

	// returning channel from request to which the topic handler writes the results
	return request.ReturnedPayload, nil
}

// connectToTCPServer function to connect to tcp server (node)
// returns connHandler channel, to which incoming parsed data is pushed
func (client *Client)connectToTCPServer(port int, errorStream chan error) (net.Conn, chan rcfUtil.Smsg, chan rcfUtil.Smsg, error) {
	topicContextMsgs = make(chan rcfUtil.Smsg)
	serviceContextMsgs = make(chan rcfUtil.Smsg)
	conn, err := net.Dial("tcp4", ":"+strconv.Itoa(port))
	if err != nil {
		return conn, topicContextMsgs, serviceContextMsgs, err
	} else {
		go client.connHandler(conn, topicContextMsgs, serviceContextMsgs, errorStream)
	}

	return conn, topicContextMsgs, serviceContextMsgs, nil
}

// NodeOpenConn initiates comm channels for handler and start them
// returns client struct which defines relevant information for the interface functions to work
func New(nodeID int, errorStream chan error) (Client, error) {
	client := new(Client)

	conn, topicContextMsgs, serviceContextMsgs, err := client.connectToTCPServer(nodeID, errorStream)
	if err != nil {
		return *client, err
	}

	topicContextRequests := make(chan dataRequest)
	serviceContextRequests := make(chan dataRequest)

	go client.topicHandler(topicContextMsgs, topicContextRequests)
	go client.serviceHandler(serviceContextMsgs, serviceContextRequests)

	client.Conn = conn
	client.clientWriteRequestCh = make(chan []byte)
	client.TopicContextRequests = topicContextRequests
	client.ServiceContextRequests = serviceContextRequests

	go client.clientWriteRequestHandler(errorStream)

	return *client, nil
}

// TopicPublishRawData pushes raw byte slice msg to topic msg stack
func (client *Client)TopicPublishData(topicName string, data []byte) error {
	encodingMsg := new(rcfUtil.Smsg)
	encodingMsg.Type = "topic"
	encodingMsg.Name = topicName
	encodingMsg.Operation = "publish"
	encodingMsg.Payload = data
	encodedMsg, err := rcfUtil.EncodeMsg(encodingMsg)
	if err != nil {
		return err
	}
	client.clientWriteRequestCh <- encodedMsg
	return nil
}

// ActionExec executes action
func (client *Client)ActionExec(actionName string, params []byte) error {
	encodingMsg := new(rcfUtil.Smsg)
	encodingMsg.Type = "action"
	encodingMsg.Name = actionName
	encodingMsg.Operation = "exec"
	encodingMsg.Payload = params
	encodedMsg, err := rcfUtil.EncodeMsg(encodingMsg)
	if err != nil {
		return err
	}
	client.clientWriteRequestCh <- encodedMsg
	encodedMsg = nil
	return nil
}

// TopicCreate creates new action on node
func (client *Client)TopicCreate(topicName string) error {
	encodingMsg := new(rcfUtil.Smsg)
	encodingMsg.Type = "topic"
	encodingMsg.Name = topicName
	encodingMsg.Operation = "create"
	encodingMsg.Payload = []byte{}
	encodedMsg, err := rcfUtil.EncodeMsg(encodingMsg)
	if err != nil {
		return err
	}
	client.clientWriteRequestCh <- encodedMsg
	encodingMsg = nil
	return nil
}

// TopicList lists node's topics
func (client *Client)TopicList() ([]string, error) {
	encodingMsg := new(rcfUtil.Smsg)
	encodingMsg.Type = "topic"
	encodingMsg.Name = "all"
	encodingMsg.Operation = "list"
	encodingMsg.MultiplePayload = [][]byte{}
	encodedMsg, err := rcfUtil.EncodeMsg(encodingMsg)
	if err != nil {
		return []string{}, err
	}
	client.clientWriteRequestCh <- encodedMsg
	encodingMsg = nil

	// creating request for the payload which is sent back from the node
	request := new(dataRequest)
	request.Name = "topiclist"
	request.Op = "pulltopiclist"
	request.Fulfilled = false
	request.PullOpReturnedPayload = make(chan [][]byte)
	// pushing request to topic handler where it is process
	client.TopicContextRequests <- *request

	reply := false
	payload := [][]byte{}

	// wainting for request to be processed and retrieval of payload
	for !reply {
		select {
		case liveDataRes := <-request.PullOpReturnedPayload:
			payload = liveDataRes
			reply = true
			break
		}
	}
	stringTopicNameList := make([]string, len(payload))
	for i, topicName := range payload {
		stringTopicNameList[i] = string(topicName)
	}
	request = nil
	return stringTopicNameList, nil
}
