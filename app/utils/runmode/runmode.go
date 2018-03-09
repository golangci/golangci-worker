package runmode

import "os"

func IsProduction() bool {
	return os.Getenv("GO_ENV") == "prod"
}
