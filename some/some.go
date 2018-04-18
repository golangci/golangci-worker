package some

import "log"

func withIssue() {
	log.Printf("bad format: %s", 1)
}
