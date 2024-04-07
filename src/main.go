package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/matthew-c-atu/project-audio-streamer/pkg/connect"
)

var audioSettings = &connect.AudioSettings{
	SampleRate: connect.DEFAULT_SAMPLERATE,
	Seconds:    connect.DEFAULT_SECONDS,
	BufferSize: connect.DEFAULT_BUFFERSIZE,
	Delay:      connect.DEFAULT_DELAY,
}

// TODO: Add flags for port selection

func main() {
	filename := flag.String("filename", "./music/bou-closer-ft-slay.aac", "path to the audio file")
	port := flag.String("port", "8080", "port on which to host application")
	flag.Parse()

	f, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}

	// contents is a byte slice

	stat, err := f.Stat()
	if err != nil {
		log.Fatal("Couldn't get file stats")
	}
	log.Printf("File size: %v\n", stat.Size())

	contents, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	// totalStreamedSize = 0
	connPool := connect.NewConnectionPool()

	log.Println("calling go stream...")
	go connect.Stream(connPool, contents, audioSettings)
	// Array equal to sample rate * 1s

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/aac")
		w.Header().Set("Connection", "keep-alive")
		// w.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := w.(http.Flusher)
		if !ok {
			panic("expected http.ResponseWriter to be an http.Flusher")
		}

		connection := &connect.Connection{BufferChannel: make(chan []byte), Buffer: make([]byte, audioSettings.BufferSize)}
		connPool.AddConnection(connection)
		log.Printf("%s has connected to the audio stream\n", r.Host)

		for {
			buf := <-connection.BufferChannel
			if _, err := w.Write(buf); err != nil {
				connPool.DeleteConnection(connection)
				log.Printf("%s's connection to the audio stream has been closed\n", r.Host)
				return
			}
			flusher.Flush() // Triger "chunked" encoding
			log.Println("emptying buffer")
			clear(connection.Buffer)
		}
	})
	log.Println("Listening on port", *port, "...")
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
