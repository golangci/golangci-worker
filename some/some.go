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

func testScopeLint() {
  funcs := []func(){}
  for _, v := range []int{1, 2} {
    funcs = append(funcs, func() {
      log.Print(v)
    })
  }

  for _, f := range funcs {
    f()
  }
}
