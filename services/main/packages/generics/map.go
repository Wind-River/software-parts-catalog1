package generics

func Map[From any, To any](in []From, fn func(From) (To, error)) ([]To, error) {
	ret := make([]To, 0)
	for _, v := range in {
		t, err := fn(v)
		if err != nil {
			return nil, err
		}

		ret = append(ret, t)
	}

	return ret, nil
}
