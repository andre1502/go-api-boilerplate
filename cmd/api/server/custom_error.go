package server

type CustomError struct {
	Message string
	Err     error
}

func (c *CustomError) Error() string {
	return c.Err.Error()
}
