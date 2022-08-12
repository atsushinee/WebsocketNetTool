package main

import (
	"golang.org/x/net/websocket"
	"io"
	"log"
)

func main() {
	const url = "ws://192.168.0.116:9009"
	dial, err := websocket.Dial(url, "ws", url)
	if err != nil {
		panic(err)
	}
	msg := make([]byte, 1024*10)
	for {
		_, err := dial.Read(msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		log.Println(string(msg))
	}

}
