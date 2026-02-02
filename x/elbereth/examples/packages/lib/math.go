package lib

func Square(X interface{}) interface{} {
	return (interface{}(X).(int64) * interface{}(X).(int64))
}

func Cube(X interface{}) interface{} {
	return (interface{}(X).(int64) * interface{}((interface{}(X).(int64) * interface{}(X).(int64))).(int64))
}
