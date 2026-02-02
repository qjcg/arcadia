package main

import "fmt"

func main() {
	fmt.Println("1 + 2 * 3 = ", (any(int64(1)).(int64) + any((any(int64(2)).(int64) * any(int64(3)).(int64))).(int64)))
	fmt.Println("10 + 20 + 30 = ", (any((any(int64(10)).(int64) + any(int64(20)).(int64))).(int64) + any(int64(30)).(int64)))
	fmt.Println("5 * 5 * 5 = ", (any((any(int64(5)).(int64) * any(int64(5)).(int64))).(int64) * any(int64(5)).(int64)))
	fmt.Println("1 + 1 + 1 * 10 = ", (any((any(int64(1)).(int64) + any(int64(1)).(int64))).(int64) + any((any(int64(1)).(int64) * any(int64(10)).(int64))).(int64)))
}
