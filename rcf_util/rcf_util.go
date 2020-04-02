package rcf_util

import(
  "regexp"
  "bytes"
  "strings"
  "encoding/gob"
)

// naming convention whitelist
var naming_whitelist string = "abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 0123456789"


// node read protocol
// ><type>-<name>-<operation>-<paypload byte slice>
func Parse_node_read_protocol(data []byte) (string, string, string, []byte) {
  var ptype string
  var name string
  var operation string
  var payload []byte

  //only for parsing purposes
  data_string := string(data)
  data_delim_split := strings.SplitN(data_string, "-", 4)
  data_delim_split_byte := bytes.SplitN(data, []byte("-"), 4)

  if(len(data)>=1 && string(data_string[0])==">") && len(data_delim_split_byte) == 4 {
    ptype = Apply_naming_conv(data_delim_split[0])
    name = Apply_naming_conv(data_delim_split[1])
    operation = data_delim_split[2]
    payload = data_delim_split_byte[3]
  }
  return ptype, name, operation, payload
}

// client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
func Topic_parse_client_read_payload(data []byte, topic_name string) []byte {
  var payload []byte

  //only for parsing purposes
  data_string := string(data)
  if(len(data)>=1) {
    // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
    if strings.Split(data_string, "-")[0] == ">topic" && strings.Split(data_string, "-")[1] == topic_name {
      var delim_index int
      for i,split_elem := range strings.Split(data_string, "-") {
        delim_index += len(split_elem)+1
        // 3 equals the amount of delimiters(-) used to the payload in the client read protocol
        if i == 3 {
          break
        }
      }
      payload = data[delim_index:]
    }
  }
  return payload
}

// client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
func Service_parse_client_read_payload(data []byte, service_name string) []byte {
  var payload []byte

  //only for parsing purposes
  data_string := string(data)
  if(len(data)>=1) {
    // client read protocol ><type>-<name>-<len(msgs)>-<paypload(msgs)>
    if strings.Split(data_string, "-")[0] == ">service" && strings.Split(data_string, "-")[1] == service_name {
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
