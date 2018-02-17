package environments

type Golang struct {
	gopath string
}

func NewGolang(gopath string) *Golang {
	return &Golang{
		gopath: gopath,
	}
}

func (g Golang) Setup(es EnvSettable) {
	es.SetEnv("GOPATH", g.gopath)
	es.SetEnv("CGO_ENABLED", "0") // Don't have gcc in heroku env, TODO: install it if needed
}
