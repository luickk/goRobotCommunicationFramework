package rcf_util

import(
  "net"
  "regexp"
  "bytes"
  "encoding/gob"
)

// naming convention whitelist
var naming_whitelist string = "abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 0123456789"

// returns first key from map
// non generic -> for string string map
func Get_first_map_key_ss(m map[string][]byte) string {
    for k := range m {
        return k
    }
    return ""
}

// returns first key from map
// non generic -> for net.Conn string map
func Get_first_map_key_cs(m map[net.Conn]string) net.Conn {
  var c net.Conn
  for k := range m {
    c = k
  }
  return c
}


// returns first key from map
// non generic -> for string net.Conn map
func Get_first_map_key_sc(m map[string]net.Conn) string {
  var c string
  for k := range m {
    c = k
  }
  return c
}

// removes last character from string
func Trim_suffix(input string) string{
  return input[:len(input)-1]
}


// removes last character from byte slice
func Trim_b_suffix_byte(input []byte) []byte{
  return input[:len(input)-1]
}

// removes last character from byte slice
func Trim_b_prefix_byte(input []byte) []byte{
  return input[1:]
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

func Topics_contains_topic(imap map[string][][]byte, key string) bool {
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

func Glob_map_decode(b *bytes.Buffer) map[string]string {
  var decodedMap map[string]string
  d := gob.NewDecoder(b)

  // Decoding the serialized data
  err := d.Decode(&decodedMap)
  if err != nil {
    panic(err)
  }
  return decodedMap
}
