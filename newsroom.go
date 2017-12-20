package main

type Newsroom struct {
	Conf *Configuration
}

// Begin running
func (nr *Newsroom) Start() {

}

func NewNewsroom(conf *Configuration) *Newsroom {
	n := new(Newsroom)
	n.Conf = conf
	return n
}
