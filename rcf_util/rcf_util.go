package rcf_util

import(
  "regexp"
  "bytes"
  "strings"
  "encoding/gob"
)

// naming convention whitelist
var naming_whitelist string = "abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 0123456789"

// client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
func Topic_parse_client_read_protocol(data []byte, topic_name string) []byte {
  var payload []byte

  //only for parsing purposes
  data_string := string(data)
  if(len(data)>=1) {
    // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
    if strings.Split(data_string, "-")[0] == ">topic" && strings.Split(data_string, "-")[1] == topic_name {
      last_del_index := strings.LastIndex(data_string, "-")
      payload = data[last_del_index+1:]
    }
  }
  return payload
}

// applies naming conventions for rcf names
func Apply_naming_conv(input_str string) string {
    reg := regexp.MustCompile("[^"+naming_whitelist+" ]+")
    topic_name_esc := reg.ReplaceAllString(input_str, "")
    return topic_name_esc
}

// requires same len slices
// compare two slices elements, return if slices are not equal
func Compare_slice(s1 []string, s2 []string) bool {
  if len(s1) != len(s2) { return false }
  for i, v := range s1 { if v != s2[i] { return false } }
  return true
}

func Topics_contain_topic(imap map[string][][]byte, key string) bool {
  if _, ok := imap[key]; ok {
    return true
  }
  return false
}

func Glob_map_encode(m map[string]string) *bytes.Buffer {
  b := new(bytes.Buffer)
  e := gob.NewEncoder(b)

  // Encoding the map
  err := e.Encode(m)
  if err != nil {
    panic(err)
  }
  return b
}

func Glob_map_decode(encoded_map []byte) map[string]string {
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
