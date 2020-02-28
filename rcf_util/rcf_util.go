package rcf_util

import(
  "regexp"
)

// naming convention whitelist
var naming_whitelist string = "abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// returns first key from map
// non generic
func Get_first_map_key(m map[string]string) string {
    for k := range m {
        return k
    }
    return ""
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
