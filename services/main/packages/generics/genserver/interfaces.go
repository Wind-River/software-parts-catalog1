package genserver

type WorkMessage[E any] struct {
	Work          E
	ReturnChannel chan error
}

type Server[E any] interface {
	Run() (chan WorkMessage[E], error)
	Listen(chan WorkMessage[E]) error
}
