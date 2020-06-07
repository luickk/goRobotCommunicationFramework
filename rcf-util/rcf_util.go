package rcf_util

import(
  "regexp"
  "bytes"
  "strconv"
  "strings"
  "encoding/gob"
  "math/rand"
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
  split := strings.Split(data, ",")
  if(len(split) == 2) {
    id, _ := strconv.Atoi(split[1])
    name := split[0]
    if id >= 0 && name != "" {
      return name, id 
    }
  }
  return "err", 0
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

func GenRandomIntId() int {
  pullReqId := rand.Intn(1000000000) 
  if pullReqId == 0 || pullReqId == 2 {
    pullReqId = rand.Intn(100000000)  
  }
  return pullReqId
}