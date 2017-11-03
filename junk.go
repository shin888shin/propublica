package main

import (
  "fmt"
  "strings"
)

func main() {
  fmt.Println("i am junk")
  x := "aaa"
  b := "bbb"
  fmt.Println(x + b)
  c := "aaa nnn"
  cc := strings.Replace(c, " ", "", -1)
  fmt.Println(cc)
}