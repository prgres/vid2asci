package app

func Before(debug bool) error {
	return logger(debug)
}
