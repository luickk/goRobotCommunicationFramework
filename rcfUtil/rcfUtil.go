/*
Package rcfutil implements basic parsing and de-encoding for rcf_node & rcf_node_client
*/
package rcfutil

import (
	"bytes"
	"encoding/gob"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

// naming convention whitelist
// every topic, action, service name is compared to that list. Characters which conflict with the protocl are removed
var namingSchemeWhitelist string = "abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 0123456789"

// basic logger declarations
// loggers are initiated by node or client
var (
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
)

// ParseNodeReadProtocol incoming data from the node client
// it return all protocol elements. type, name, operation, payload
func ParseNodeReadProtocol(data []byte) (string, string, string, int, []byte) {
	InfoLogger.Println("ParseNodeReadProtocol called")
	var ptype string
	var name string
	var operation string
	var payloadLen int
	var payload []byte

	//only for parsing purposes
	dataString := string(data)
	dataDelimSplit := strings.SplitN(dataString, "-", 5)
	dataDelimSplitByte := bytes.SplitN(data, []byte("-"), 5)

	if (len(data) >= 1 && string(dataString[0]) == ">") && len(dataDelimSplitByte) == 5 {
		ptype = ApplyNamingConv(dataDelimSplit[0])
		name = dataDelimSplit[1]
		operation = dataDelimSplit[2]
		pLen, err := strconv.Atoi(string(dataDelimSplitByte[3]))
		if err != nil {
			payloadLen = 0
		} else {
			payloadLen = pLen
		}
		payload = dataDelimSplitByte[4]
		InfoLogger.Println("ParseNodeReadProtocol data parsed")
	}
	return ptype, name, operation, payloadLen, payload
}

// SplitServiceToNameID splits name field of the protocol to name and id.
// this is recquired because the id is "encoded" in the name
// <name,id>
// returns name, id
func SplitServiceToNameID(data string) (string, int) {
	InfoLogger.Println("SplitServiceToNameId called")
	split := strings.Split(data, ",")
	if len(split) == 2 {
		id, err := strconv.Atoi(split[1])
		if err != nil {
			WarningLogger.Println("SplitServiceToNameId conversion err")
		} else {
			name := split[0]
			if id >= 0 && name != "" {
				return name, id
				InfoLogger.Println("SplitServiceToNameId name,id split")
			}
		}
	}
	InfoLogger.Println("SplitServiceToNameId could not split name")
	return "err", 0
}

// ApplyNamingConv applies naming conventions for rcf names
// returns corrected name
func ApplyNamingConv(inputStr string) string {
	reg := regexp.MustCompile("[^" + namingSchemeWhitelist + " ]+")
	topicNameEsc := reg.ReplaceAllString(inputStr, "")
	InfoLogger.Println("ApplyNamingConv called")
	return topicNameEsc
}

// CompareSlice compares two slices for equality
// slices must be of same length
// returns false if slices are not equal
func CompareSlice(s1 []string, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v := range s1 {
		if v != s2[i] {
			return false
		}
	}
	InfoLogger.Println("CompareSlice called")
	return true
}

// TopicsContainTopic checks if the topics map contains a certain topic(name)
// returns false if topic(name) is not included in the list
func TopicsContainTopic(imap map[string][][]byte, key string) bool {
	if _, ok := imap[key]; ok {
		return true
	}
	InfoLogger.Println("TopicsContainTopic called")
	return false
}

// GlobMapEncode serializes any map to byte buffer
// returnes serialized map as buffer
func GlobMapEncode(m map[string]string) (*bytes.Buffer, error) {
	InfoLogger.Println("GlobMapEncode called")
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)

	// Encoding the map
	err := e.Encode(m)

	return b, err
}

// GlobMapDecode decodes serialized glob map byte array to valid map if k:string,v:string map
// returns k:string,v:string map, according to the protocol
func GlobMapDecode(encodedMap []byte, s string) (map[string]string, error) {
	InfoLogger.Println("GlobMapDecode called")
	b := bytes.NewBuffer(make([]byte, 0, len(encodedMap)))
	b.Write(encodedMap)
	var decodedMap map[string]string
	d := gob.NewDecoder(b)
	// Decoding the serialized data
	err := d.Decode(&decodedMap)
	return decodedMap, err
}

// GenRandomIntID generates random id
// returns generated random id
func GenRandomIntID() int {
	InfoLogger.Println("GenRandomIntID called")
	pullReqID := rand.Intn(1000000000)
	if pullReqID == 0 || pullReqID == 2 {
		pullReqID = rand.Intn(100000000)
	}
	return pullReqID
}
