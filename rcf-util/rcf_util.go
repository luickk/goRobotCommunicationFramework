package rcf_util

import(
  "regexp"
  "bytes"
  "strconv"
  "strings"
  "encoding/gob"
)

// naming convention whitelist
var namingSchemeWhitelist string = "abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 0123456789"

// node read protocol
// ><type>-<name>-<operation>-<paypload byte slice>
func ParseNodeReadProtocol(data []byte) (string, string, string, []byte) {
  var ptype string
  var name string
  var operation string
  var payload []byte

  //only for parsing purposes
  dataString := string(data)
  dataDelimSplit := strings.SplitN(dataString, "-", 4)
  dataDelimSplitByte := bytes.SplitN(data, []byte("-"), 4)

  if(len(data)>=1 && string(dataString[0])==">") && len(dataDelimSplitByte) == 4 {
    ptype = ApplyNamingConv(dataDelimSplit[0])
    name = dataDelimSplit[1]
    operation = dataDelimSplit[2]
    payload = dataDelimSplitByte[3]
  }
  return ptype, name, operation, payload
}

// node read protocol extends ids for services
// <name,id>
// returns name, id
func SplitServiceToNameId(data string) (string, int) {
  split := strings.Split(data, "&")
  if(len(split) == 2) {
    id, _ := strconv.Atoi(split[1])
    name := split[0]
    if id >= 0 && name != "" {
      return name, id 
    }
  }
  return "err", 0
}

// client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
func TopicParseClientReadPayload(data []byte, topic_name string) []byte {
  var payload []byte

  //only for parsing purposes
  dataString := string(data)
  if(len(data)>=1) {
    // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
    if strings.Split(dataString, "-")[0] == ">topic" && strings.Split(dataString, "-")[1] == topic_name {
      payload = bytes.SplitN(data, []byte("-"), 4)[3]
    }
  }
  return payload
}

// client read protocol ><type>-<name>-<len(msg)>-<paypload(msgs)>
func ServiceParseClientReadPayload(data []byte, serviceName string, serviceId int) []byte {
  var payload []byte

  //only for parsing purposes
  dataString := string(data)
  //print(dataString)
  if(len(data)>=1) {
    // client read protocol ><type>-<name>-<serviceId>-<paypload(msgs)>
    // println("Data String: " + dataString)
    splitData := strings.Split(dataString, "-")
    if len(splitData) >= 2 {
      msgType := splitData[0]
      msgServiceName := splitData[1]
      msgServiceOnlyName, msgServiceId := SplitServiceToNameId(msgServiceName)
      // println("Found service reply "+msgType + msgServiceName)
      if msgType == ">service" && msgServiceOnlyName == serviceName && msgServiceId == serviceId {
        payload = bytes.SplitN(data, []byte("-"), 4)[3]
        println("Service Id: "+strconv.Itoa(msgServiceId))
        println("Service Name: "+ msgServiceName)
      }
    }
  }
  return payload
}

// applies naming conventions for rcf names
func ApplyNamingConv(input_str string) string {
    reg := regexp.MustCompile("[^"+namingSchemeWhitelist+" ]+")
    topic_name_esc := reg.ReplaceAllString(input_str, "")
    return topic_name_esc
}

// requires same len slices
// compare two slices elements, return if slices are not equal
func CompareSlice(s1 []string, s2 []string) bool {
  if len(s1) != len(s2) { return false }
  for i, v := range s1 { if v != s2[i] { return false } }
  return true
}

func TopicsContainTopic(imap map[string][][]byte, key string) bool {
  if _, ok := imap[key]; ok {
    return true
  }
  return false
}

func GlobMapEncode(m map[string]string) *bytes.Buffer {
  b := new(bytes.Buffer)
  e := gob.NewEncoder(b)

  // Encoding the map
  err := e.Encode(m)
  if err != nil {
    panic(err)
  }
  return b
}

func GlobMapDecode(encoded_map []byte) map[string]string {
  b := bytes.NewBuffer(make([]byte,0,len(encoded_map)))
  b.Write(encoded_map)
  var decodedMap map[string]string
  d := gob.NewDecoder(b)
  // Decoding the serialized data
  err := d.Decode(&decodedMap)
  if err != nil {
    panic(err)
  }
  return decodedMap
}
