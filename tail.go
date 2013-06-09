package tail

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.exp/fsnotify"
	"io"
	"log"
	"os"
)

var (
	watcher   *fsnotify.Watcher
	c         chan Update
	watchlist map[string]watched

	debug = true
)

type watched struct {
	file   *os.File
	reader *bufio.Reader
	buf    bytes.Buffer
}

type Update struct {
	File     string
	Contents []byte
}

func init() {
	var err error

	if watcher, err = fsnotify.NewWatcher(); err != nil {
		log.Fatal(err)
	}

	c = make(chan Update)
	watchlist = make(map[string]watched)

	go listen()
}

// Joka file tarvitsee oman bufferinsa :( ja tieto filen nimestä kun lähettää
func listen() {
	var (
		ok    bool
		err   error
		b     []byte
		event *fsnotify.FileEvent

		filename string
		file     watched
	)

listenloop:
	for {
		select {
		case event, ok = <-watcher.Event:
			if !ok {
				break listenloop
			}
			filename = event.Name
			if file, ok = watchlist[filename]; !ok {
				log.Printf("Tail: Event from %s, file not in watchlist", filename)
				continue
				// NOTE: try to open the file in question?
			}
			if b, err = file.reader.ReadBytes('\n'); err != nil {
				if err == io.EOF {
					file.buf.Write(b)
					continue
				} else {
					log.Printf("Tail: Error while reading buffer - %s", err.Error())
					continue
				}
			}
			file.buf.Write(b)
			if debug {
				log.Printf("Sent: %s", string(b))
			}
			if len(b) > 0 {
				c <- Update{filename, file.buf.Bytes()}
				file.buf.Reset()
			}
		}
	}

}

func Connect() (event <-chan Update, err <-chan error) {
	return (<-chan Update)(c), (<-chan error)(watcher.Error)
}

// TODO: Return n previous lines of each file when Added
func Add(path string) (err error) {
	var f *os.File

	if f, err = os.Open(path); err != nil {
		return
	}
	watchlist[path] = watched{file: f, reader: bufio.NewReader(f)}

	return watcher.WatchFlags(path, fsnotify.FSN_MODIFY)
}

func Remove(path string) (err error) {
	var (
		f  watched
		ok bool
	)

	if f, ok = watchlist[path]; !ok {
		log.Printf("Tail: Path %s not in watchlist", path)
	} else {
		f.file.Close()
		delete(watchlist, path)
	}

	return watcher.RemoveWatch(path)
}
