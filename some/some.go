package some

import "log"

func withIssue() {
	log.Printf("bad format: %s", 1)
}

func withAnotherIssue() {
  a := 1
  if a != 0 {
    return
  } else {
    panic(a)
  }
}
