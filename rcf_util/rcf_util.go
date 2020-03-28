package rcf_util

import(
  "regexp"
  "bytes"
  "encoding/gob"
)

// naming convention whitelist
var naming_whitelist string = "abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 0123456789"

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
