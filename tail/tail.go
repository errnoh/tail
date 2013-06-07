// FOR TESTING ONLY, WIP
package main

import (
	"fmt"
	"github.com/errnoh/tail"
	"io/ioutil"
	"os"
)

func main() {
	var (
		f      *os.File
		err    error
		events <-chan []byte
		errors <-chan error
	)

	events, errors = tail.Connect()

	if f, err = ioutil.TempFile("", "append"); err != nil {
		fmt.Println(err)
		return
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if err = tail.Add(f.Name()); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Listening for file", f.Name())

	for {
		select {
		case e := <-events:
			if string(e) == "close\n" {
				return
			}
			fmt.Println("Received:",string(e),[]byte(e))
		case err = <-errors:
			fmt.Println(err)
			return
		}
	}
}
