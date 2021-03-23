package main

import (
	"fmt"
)

func main() {
	z := NewZipList()

	z.Push(-1<<31 - 20)
	z.Push(1<<15 + 10)
	z.Push("Hello")
	z.Push("肖何子建")
	z.Push("World")


	fmt.Println(z.ZLBytes())
	fmt.Println(z.ZLTail())
	fmt.Println(z.ZLLen())

	arr := z.All()
	fmt.Println(arr)
	fmt.Println(z.Index(2147483658))
	fmt.Println(z.Index(-2147483668))
	fmt.Println(z.Index("肖何子建"))

}
