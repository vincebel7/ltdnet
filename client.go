package main

import( "fmt"
	"encoding/json"
//	"os"
	"io/ioutil"
)

type network struct {
	name []string
}

func main() {
	fmt.Println("test")

	myint, _ := json.Marshal(3)
	fmt.Println(string(myint))
}
