// FOR TESTING ONLY, WIP
package main

import (
	"flag"
	"fmt"
	"github.com/errnoh/tail"
)

func main() {
	var (
		err    error
		events <-chan tail.Update
		errors <-chan error
	)

	flag.Parse()

	events, errors = tail.Connect()

	for _, filename := range flag.Args() {
		if err = tail.Add(filename); err != nil {
			fmt.Println(err)
			return
		}
		defer tail.Remove(filename)
		fmt.Println("Listening for file", filename)
	}


	for {
		select {
		case e := <-events:
			if string(e.Contents) == "close\n" {
				return
			}
			fmt.Printf("%s: %s", e.File, string(e.Contents))
		case err = <-errors:
			fmt.Println(err)
			return
		}
	}
}
