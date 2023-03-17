package genserver

func Call[E any](server Server[E], work E) error {
	workChannel, err := server.Run()
	if err != nil {
		return err
	}

	returnChannel := make(chan error)
	workChannel <- WorkMessage[E]{
		Work:          work,
		ReturnChannel: returnChannel,
	}

	return <-returnChannel
}

func Cast[E any](server Server[E], work E) error {
	workChannel, err := server.Run()
	if err != nil {
		return err
	}

	workChannel <- WorkMessage[E]{
		Work:          work,
		ReturnChannel: nil,
	}

	return nil
}
