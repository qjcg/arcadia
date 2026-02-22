package lib

func Square(X any) any {
	return (any(X).(int64) * any(X).(int64))
}

func Cube(X any) any {
	return (any(X).(int64) * any((any(X).(int64) * any(X).(int64))).(int64))
}
