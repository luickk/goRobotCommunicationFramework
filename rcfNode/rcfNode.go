/*
Robot Communication Framework

The RCF is a framework for data distribution, which the most essential part of an autonomous platform.
It is very similar to ROS but without packages, the C/C++ complexity overhead while still maintaining speed and safe
thread/ lang. standards, thanks to the go lang.

*/

package rcfnode

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"rcf/rcfUtil"
  "rcf/tools"
)

// topicCapacity defines the amount of msgs a single topic queue "stores" before they are overwritten
var topicCapacity = 5

// tcpConnBuffer defines the buffer size of the TCP conn reader
var tcpConnBuffer = 1024

// basic logger declarations
var (
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
)

// describes a single client write request
// exists to forward client write requests form the handlers to the clientWrite handler via a channel
type clientWriteRequest struct {
	receivingClient net.Conn
	msg             []byte
}

// describes a single msgs
// exists to forward requests form the connection handler to the topic handler via a channel in pull,push/ subscribe events
type topicMsg struct {
	topicName string
	msg       []byte
}

// describes a pull request from a client
// exists to send requests form the connection handler to the topic handler via a channel
type topicPullReq struct {
	conn      net.Conn
	topicName string
	nmsg      int
}

// describes a single connection which is subscribed to a certain topic
// exists to forward subscribe requests form the connection handler to the topic handler via a channel
type topicListenerConn struct {
	listeningConn net.Conn
	topicName     string
}

// defines a single action with the executed function and its name
// exists to forward created actions from interface function to action handler
type action struct {
	actionName     string
	actionFunction actionFn
}

// defines a single service with the executed function and its name
// exists to forward created services from interface function to service handler
type service struct {
	serviceName     string
	serviceFunction serviceFn
}

// defines an services call request
// contains the connection it is called by such as the called services name and the call params
type serviceExec struct {
	serviceName     string
	serviceCallConn net.Conn
	params          []byte
}

// defines as action call request
// contains the connection it is called by such as the called action name and the call params
type actionExec struct {
	actionName string
	params     []byte
}

// Node struct contains all Node attributes and data maps
type Node struct {
	// id or port of node
	id int

	// clientWriteRequestCh is the channel alle write requests are written to
	clientWriteRequestCh chan *clientWriteRequest

	// topic map is handled by the topic handler and contains all topics data
	// key: topic name, value: stack slice
	topics map[string][][]byte

	// contains topic msgs which should be pushed to a topic
	// handled by topic handler
	topicPushCh chan *topicMsg

	// contains topics which should be created
	// handled by topic handler
	topicCreateCh chan string

	// contains topic subscribe requests which should be processed by the handler handler
	topicListenerConnCh chan *topicListenerConn

	// contains all topic pull request which should be processed by the topic handler
	topicPullCh chan *topicPullReq

	// contains all active topic listeners
	// handled by topic handler
	topicListenerConns []*topicListenerConn

	// actions map is handled by the action handler and contains all actions
	// key: action name, value: action function
	actions map[string]*actionFn

	// contains actions which should be created
	// handled by action handler
	actionCreateCh chan *action

	// contains actions which should be executed
	// handled by action handler
	actionExecCh chan *actionExec

	// services map is handled by the service handler and contains all services
	// key: service name, value: service function
	services map[string]*serviceFn

	// cotains services which should be created
	// handled by service handler
	serviceCreateCh chan *service

	// contains service s which should be executed
	// handled by service handler
	serviceExecCh chan *serviceExec
}

// general action function type
type actionFn func(params []byte, nodeInstance Node)

// general service function type
type serviceFn func(params []byte, nodeInstance Node) []byte

// handles every incoming instructions from the client and executes them
func (node *Node)handleConnection(conn net.Conn) {
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
		clientWriteRequest := new(clientWriteRequest)
		encodingMsg := new(rcfUtil.Smsg)
		switch decodedMsg.Type {
		case "topic":
			switch decodedMsg.Operation {
			case "publish":
				node.TopicPublishData(decodedMsg.Name, decodedMsg.Payload)
			case "pull":
				// handles pull reueqst from client
				node.TopicPullData(conn, decodedMsg.Name, len(decodedMsg.Payload))
			case "subscribe":
				node.TopicAddListenerConn(decodedMsg.Name, conn)
			case "create":
				node.TopicCreate(decodedMsg.Name)
			case "list":
				encodingMsg
				clientWriteRequest.receivingClient = conn

				encodingMsg.Type = "topic"
				encodingMsg.Name = "topiclist"
				encodingMsg.Operation = "pullinfo"
				encodingMsg.payload = decodedMsg.Payload
				clientWriteRequest.msg = rcfUtil.EncodeMsg(encodingMsg)

				node.clientWriteRequestCh <- clientWriteRequest
			}
		case "action":
			if decodedMsg.Operation == "exec" {
				InfoLogger.Println("handleConnection action execed")
				node.ActionExec(decodedMsg.Name, decodedMsg.Payload)
			}
		case "service":
			if decodedMsg.Operation == "exec" {
				InfoLogger.Println("handleConnection service execed")
				node.ServiceExec(conn, decodedMsg.Name, decodedMsg.Payload)
			}
		}
		byteData = []byte{}
	}
}

// clientWriteRequestHandler handles all write request to clients
func (node *Node)clientWriteRequestHandler() {
	InfoLogger.Println("writeHandler started")
	for {
		select {
		case writeRequest := <-node.clientWriteRequestCh:
			writeRequest.receivingClient.Write(append(writeRequest.msg, []byte{0x00}))
		}
	}
}

// topicHandler handles all memory critical read, write operations to the topics map and reduces the topic maps slices to given max length
// as well as pull,pus,sub operations from the client
func (node *Node)topicHandler() {
	encodingMsg := new(rcfUtil.Smsg)
	clientWriteRequest := new(clientWriteRequest)
	encodingMsg := new(rcfUtil.Smsg)
	for {
		select {
		case topicListener := <-node.topicListenerConnCh:
			InfoLogger.Println("topicHandler listener added")
			// appends new client listener to active listener slice
			node.topicListenerConns = append(node.topicListenerConns, topicListener)
		case pullRequest := <-node.topicPullCh:
			InfoLogger.Println("topicHandler data pulled")
			var byteData [][]byte
			// parses the non unique topic name from the instruction
			topicOnlyName, _ := rcfUtil.SplitServiceToNameID(pullRequest.topicName)

			if pullRequest.nmsg >= len(node.topics[topicOnlyName]) {
				byteData = node.topics[topicOnlyName]
			} else {
				byteData = node.topics[topicOnlyName][:pullRequest.nmsg]
			}
			if len(byteData) >= 1 {
				if pullRequest.nmsg <= 1 {
					clientWriteRequest.receivingClient = pullRequest.conn

					encodingMsg.Type = "topic"
					encodingMsg.Name = pullRequest.topicName
					encodingMsg.Operation = "pull"
					encodingMsg.payload = byteData[0]
					clientWriteRequest.msg = rcfUtil.EncodeMsg(encodingMsg)

					node.clientWriteRequestCh <- clientWriteRequest
				} else {
					tdata := append(bytes.Join(byteData, []byte("")), []byte("\r")...)
					lengths := []int{}
					for _, data := range byteData {
						lengths = append(lengths, len(data))
					}
					clientWriteRequest.receivingClient = pullRequest.conn

					encodingMsg.Type = "topic"
					encodingMsg.Name = pullRequest.topicName
					encodingMsg.Operation = "pull"
					encodingMsg.payload = tdata
					clientWriteRequest.msg = rcfUtil.EncodeMsg(encodingMsg)

					node.clientWriteRequestCh <- clientWriteRequest
				}
			} else {
				clientWriteRequest.receivingClient = pullRequest.conn

				encodingMsg.Type = "topic"
				encodingMsg.Name = pullRequest.topicName
				encodingMsg.Operation = "pull"
				encodingMsg.payload = []byte{}
				clientWriteRequest.msg = rcfUtil.EncodeMsg(encodingMsg)

				node.clientWriteRequestCh <- clientWriteRequest
			}
		case topicMsg := <-node.topicPushCh:
			InfoLogger.Println("topicHandler data pushed")
			if rcfUtil.TopicsContainTopic(node.topics, topicMsg.topicName) {

				node.topics[topicMsg.topicName] = append(node.topics[topicMsg.topicName], topicMsg.msg)

				// check if topic exceeds topic cap limits
				if len(node.topics[topicMsg.topicName]) > topicCapacity {

					topicOverhead := len(node.topics[topicMsg.topicName]) - topicCapacity
					// slicing size of slice to right size‚
					node.topics[topicMsg.topicName] = node.topics[topicMsg.topicName][topicOverhead:]

				}

				// check if topic, which data is pushed to, has a listening conn
				for _, topicListener := range node.topicListenerConns {

					topicOnlyName, _ := rcfUtil.SplitServiceToNameID(topicListener.topicName)

					if topicOnlyName == topicMsg.topicName && len(topicMsg.msg) != 0 {
						clientWriteRequest.receivingClient = topicListener.listeningConn

						encodingMsg.Type = "topic"
						encodingMsg.Name = pullRequest.topicName
						encodingMsg.Operation = "sub"
						encodingMsg.payload = []byte(topicMsg.msg)
						clientWriteRequest.msg = rcfUtil.EncodeMsg(encodingMsg)

						node.clientWriteRequestCh <- clientWriteRequest
					}
				}
			}

		case topicCreateName := <-node.topicCreateCh:
			if !rcfUtil.TopicsContainTopic(node.topics, topicCreateName) {
				node.topics[topicCreateName] = [][]byte{}
				InfoLogger.Println("topicHandler topic created")
			}
		}
	}
}

// handles all memory critical read, write operations to the actions map as well as create, execution instruction operations from the client
func (node *Node)actionHandler() {
	InfoLogger.Println("actionHandler started")
	for {
		select {
		case action := <-nodeInstance.actionCreateCh:
			nodeInstance.actions[action.actionName] = &action.actionFunction
			InfoLogger.Println("actionHandler action created")
		case actionExec := <-nodeInstance.actionExecCh:
			InfoLogger.Println("actionHandler action execed")
			if _, ok := nodeInstance.actions[actionExec.actionName]; ok {
				actionFunc := *nodeInstance.actions[actionExec.actionName]
				go actionFunc(actionExec.params, nodeInstance)
			} else {
				InfoLogger.Println("actionHandler action execed not found")
			}
		}
	}
}

// handles all memory critical read, write operations to the services map as well as create, execution instruction operations from the client
func (nodeInstance *Node)serviceHandler() {
	clientWriteRequest := new(clientWriteRequest)
	encodingMsg := new(rcfUtil.Smsg)
	for {
		select {
		case service := <-nodeInstance.serviceCreateCh:
			nodeInstance.services[service.serviceName] = &service.serviceFunction
			InfoLogger.Println("serviceHandler service created")
		case serviceExec := <-nodeInstance.serviceExecCh:
			InfoLogger.Println("serviceHandler service execed(queued)")
			serviceOnlyName, _ := rcfUtil.SplitServiceToNameID(serviceExec.serviceName)
			if _, ok := nodeInstance.services[serviceOnlyName]; ok {
				serviceFn := *nodeInstance.services[serviceOnlyName]
				go func() {
					serviceResult := serviceFn(serviceExec.params, nodeInstance)
					clientWriteRequest.receivingClient = serviceExec.serviceCallConn

					encodingMsg.Type = "service"
					encodingMsg.Name = serviceExec.serviceName
					encodingMsg.Operation = "called"
					encodingMsg.payload = serviceResult
					clientWriteRequest.msg = rcfUtil.EncodeMsg(encodingMsg)

					nodeInstance.clientWriteRequestCh <- clientWriteRequest

					InfoLogger.Println("serviceHandler service returned and wrote payload")
				}()
			} else {
				clientWriteRequest.receivingClient = serviceExec.serviceCallConn


				encodingMsg.Type = "service"
				encodingMsg.Name = serviceExec.serviceName
				encodingMsg.Operation = "called"
				encodingMsg.payload = []byte{}
				clientWriteRequest.msg = rcfUtil.EncodeMsg(encodingMsg)

				nodeInstance.clientWriteRequestCh <- clientWriteRequest
			}
		}
	}
}

// Create initiates node instance and initializes all channels, maps
func New(nodeID int) Node {

	clientWriteRequestCh := make(chan *clientWriteRequest, 100)

	// key: topic name, value: stack slice
	topics := make(map[string][][]byte)

	topicPushCh := make(chan *topicMsg)

	topicCreateCh := make(chan string)

	topicListenerConnCh := make(chan *topicListenerConn)

	topicPullCh := make(chan *topicPullReq)

	topicListenerConns := make([]*topicListenerConn, 0)

	// action map with first key(action name) value(anon action func) pair
	actions := make(map[string]*actionFn)

	actionCreateCh := make(chan *action)

	actionExecCh := make(chan *actionExec)

	services := make(map[string]*serviceFn)

	serviceCreateCh := make(chan *service)

	serviceExecCh := make(chan *serviceExec)
	node := Node{nodeID, clientWriteRequestCh, topics, topicPushCh, topicCreateCh, topicListenerConnCh, topicPullCh, topicListenerConns, actions, actionCreateCh, actionExecCh, services, serviceCreateCh, serviceExecCh}
	node.init()
	return node
}

// Init initiates node with given id
// returns initiated node instance to enable direct service and topic operations
func (node *Node)init() {
	InfoLogger = log.New(os.Stdout, "[NODE] INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "[NODE] WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "[NODE] ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// disableing debug information
	InfoLogger.SetOutput(ioutil.Discard)
	// ErrorLogger.SetOutput(ioutil.Discard)
	// WarningLogger.SetOutput(ioutil.Discard)

	// initiating basic loggers
	rcfUtil.InfoLogger = InfoLogger
	rcfUtil.WarningLogger = WarningLogger
	rcfUtil.ErrorLogger = ErrorLogger

	// starting all handlers
	go node.topicHandler()
	go node.actionHandler()
	go node.serviceHandler()
	go node.clientWriteRequestHandler()

	InfoLogger.Println("Init handlers routine started")
	var port string = ":" + strconv.Itoa(node.id)
	l, err := net.Listen("tcp4", port)

	if err != nil {
		ErrorLogger.Println(err)
	}
	go func() {
		tools.Dump()
		defer l.Close()

		for {
			conn, err := l.Accept()
			if err != nil {
				ErrorLogger.Println("Init client conn accept error")
				ErrorLogger.Println(err)
			}
			go node.handleConnection(conn)
			InfoLogger.Println("Init new conn handler routine started")
		}
	}()
}

// NodeHalt pauses node
func NodeHalt() {
	InfoLogger.Println("NodeHalt called")
	for {
		time.Sleep(1 * time.Second)
	}
}

// TopicAddListenerConn adds new subscribe request to active listener conn slice
func (node *Node)TopicAddListenerConn(topicName string, conn net.Conn) {
	InfoLogger.Println("TopicAddListenerConn called")
	// create subscribe request
	topicListenerConn := new(topicListenerConn)
	topicListenerConn.topicName = topicName
	topicListenerConn.listeningConn = conn
	// sending subscribe request
	node.topicListenerConnCh <- topicListenerConn
	topicListenerConn = nil
}

// NodeListTopics returns all topic names
func NodeListTopics(node Node) []string {
	InfoLogger.Println("NodeListTopics called")
	keys := make([]string, 0, len(node.topics))
	for k := range node.topics {
		keys = append(keys, k)
	}
	if len(node.topics) > 0 {
		return keys
	} else if len(node.topics) == 0 {
		return []string{"none"}
	}
	return []string{"none"}
}

// TopicPullData creates pull request and sends it to the topic handler
// topic handler processes created request and sends result to given conn
func (node *Node)TopicPullData(conn net.Conn, topicName string, nmsg int) {
	InfoLogger.Println("TopicPullData called")
	// creating pull request
	topicPullReq := new(topicPullReq)
	topicPullReq.topicName = topicName
	topicPullReq.nmsg = nmsg
	topicPullReq.conn = conn
	// sending pull request to topic handler
	node.topicPullCh <- topicPullReq
}

// TopicPublishData creates push request and sends it to the topic handler
// topic handler processes created request and adds new topic msg to given topic name
func (node *Node)TopicPublishData(topicName string, tdata []byte) {
	InfoLogger.Println("TopicPublishData called")
	topicMsg := new(topicMsg)
	topicMsg.topicName = topicName
	topicMsg.msg = tdata
	node.topicPushCh <- topicMsg
	topicMsg = nil
}

// TopicCreate creates topic create request and sends it to the topic handler
// topic handler processes created request and adds new topic to the topics map
func (node *Node)TopicCreate(topicName string) {
	InfoLogger.Println("TopicCreate called")
	topicName = rcfUtil.ApplyNamingConv(topicName)

	node.topicCreateCh <- topicName
}

// ActionCreate creates action create request and sends it to the action handler
// action handler processes created request and adds new action with given action name
func (node *Node)ActionCreate(actionName string, actionFunc actionFn) {
	actionName = rcfUtil.ApplyNamingConv(actionName)
	newAction := new(action)
	newAction.actionName = actionName
	newAction.actionFunction = actionFunc
	node.actionCreateCh <- newAction
	newAction = nil
	InfoLogger.Println("ActionCreate called")
}

// ActionExec creates actions execution request and sends it to the action handler
// action handler processes created request and executes the action
func (node *Node)ActionExec(actionName string, actionParams []byte) {
	InfoLogger.Println("ActionExec called")
	actionExec := new(actionExec)
	actionExec.actionName = actionName
	actionExec.params = actionParams
	node.actionExecCh <- actionExec
	actionExec = nil
}

// ServiceCreate creates service create request and sends it to the service handler
// service handler processes created request and adds new service with given service name
func (node *Node)ServiceCreate(serviceName string, serviceFunc serviceFn) {
	serviceName = rcfUtil.ApplyNamingConv(serviceName)
	service := new(service)
	service.serviceName = serviceName
	service.serviceFunction = serviceFunc
	node.serviceCreateCh <- service
	service = nil
	InfoLogger.Println("ServiceCreate called")
}

// ServiceExec creates service execution request and sends it to the service handler
// service handler processes created request and executes the service as well as returns the result to the given connection(client who called the service)
func (node *Node)ServiceExec(conn net.Conn, serviceName string, serviceParams []byte) {
	InfoLogger.Println("ServiceExec called")
	serviceExec := new(serviceExec)
	serviceExec.serviceName = serviceName
	serviceExec.serviceCallConn = conn
	serviceExec.params = serviceParams
	node.serviceExecCh <- serviceExec
	serviceExec = nil
}
