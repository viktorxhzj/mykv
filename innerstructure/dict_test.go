package innerstructure

import (
	"fmt"
	"testing"
)

func TestDict_Put(t *testing.T) {
	d := NewDict()

	keys := []string{"Hello", "分布式小盒子"}

	for i := range keys {
		d.Put(keys[i], keys[i])
	}

	fmt.Println(len(keys), d.Size())

	for _, k := range keys {
		fmt.Println(d.Get(k))
	}
}
