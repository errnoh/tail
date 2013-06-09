package tail

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

// defer is LIFO

func TestAppend(t *testing.T) {
	var (
		f   *os.File
		err error
	)

	if f, err = ioutil.TempFile("", "append"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if _, err = f.Write([]byte("bacon")); err != nil {
		t.Fatal(err)
	}
}

func TestEventReceive(t *testing.T) {
	var (
		f      *os.File
		err    error
		events <-chan Update
		errors <-chan error
	)

	events, errors = Connect()

	if f, err = ioutil.TempFile("", "append"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if err = Add(f.Name()); err != nil {
		t.Fatal(err)
	}

	if _, err = f.Write([]byte("bacon\n")); err != nil {
		t.Fatal(err)
	}

	go func() {
		f.Write([]byte("crispy"))
		time.Sleep(time.Millisecond * 200)
		f.Write([]byte(" bacon\n"))
	}()

	time.Sleep(time.Millisecond * 300)

	for i := 0; i < 2; i++ {
		select {
		case e := <-events:
			t.Logf("Received: %s", e)
			if string(e.Contents) != "bacon\n" && string(e.Contents) != "crispy bacon\n" {
				t.Fatalf("Wrong input, expected either \"bacon\" or \"crispy bacon\"")
			}
		case err = <-errors:
			t.Fatal(err)
		case <-time.After(time.Second):
			t.Fatal("timeout while waiting for event")
		}
	}
}
