package rcf_util

import(
  "net"
  "regexp"
)

// naming convention whitelist
var naming_whitelist string = "abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// returns first key from map
// non generic -> for string string map
func Get_first_map_key_ss(m map[string]string) string {
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
// non generic -> for net.Conn string map
func Get_first_map_key_si(m map[string]interface{}) string {
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

func Topics_contains_topic(imap map[string][]string, key string) bool {
  if _, ok := imap[key]; ok {
    return true
  }
  return false
}
