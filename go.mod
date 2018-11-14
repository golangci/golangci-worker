module github.com/golangci/golangci-worker

// +heroku goVersion go1.11
// +heroku install ./cmd/...

require (
	github.com/RichardKnop/machinery v0.0.0-20180221144734-c5e057032f00
	github.com/cenkalti/backoff v2.0.0+incompatible
	github.com/dukex/mixpanel v0.0.0-20170510165255-53bfdf679eec
	github.com/golang/mock v1.1.1
	github.com/golangci/getrepoinfo v0.0.0-20180818083854-2a0c71df2c85
	github.com/golangci/golangci-api v0.0.0-20181114200623-38113e64849c
	github.com/golangci/golangci-lint v0.0.0-20181114200623-a84578d603c7
	github.com/golangci/golangci-shared v0.0.0-20181003182622-9200811537b3
	github.com/google/go-github v0.0.0-20180123235826-b1f138353a62
	github.com/joho/godotenv v0.0.0-20180115024921-6bb08516677f
	github.com/levigross/grequests v0.0.0-20180717012718-3f841d606c5a
	github.com/pkg/errors v0.8.0
	github.com/savaki/amplitude-go v0.0.0-20160610055645-f62e3b57c0e4
	github.com/shirou/gopsutil v0.0.0-20180801053943-8048a2e9c577
	github.com/sirupsen/logrus v1.0.5
	github.com/stretchr/testify v1.2.1
	golang.org/x/oauth2 v0.0.0-20180118004544-b28fcf2b08a1
)
