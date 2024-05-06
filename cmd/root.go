/*
Copyright © 2024 NME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/matthew-c-atu/project-audio-streamer/pkg/connect"
	"github.com/spf13/cobra"
)

var audioSettings = &connect.AudioSettings{
	SampleRate: connect.DEFAULT_SAMPLERATE,
	Seconds:    connect.DEFAULT_SECONDS,
	BufferSize: connect.DEFAULT_BUFFERSIZE,
	Delay:      connect.DEFAULT_DELAY,
}

type RootCfg struct{ *cobra.Command }

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "project-audio-streamer",
	Short: "Audio streaming server application which serves audio files over HTTP",
	Long: `An audio streaming server application which serves WAV audio files over HTTP.
	TODO: Fill out more of this description
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		root := &RootCfg{cmd}
		root.serveHls()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().IntP("port", "p", 8080, "The port on which to host the server")
	rootCmd.PersistentFlags().StringP("filepath", "f", "music", "path to the audio file")
}

func (r *RootCfg) serve() {

	// TODO: Add error handling/logging
	port, _ := r.Flags().GetInt("port")

	// TODO: Add error handling/logging
	filepath, _ := r.Flags().GetString("filepath")

	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}

	stat, err := f.Stat()
	if err != nil {
		log.Fatal("Couldn't get file stats")
	}
	log.Printf("File size: %v\n", stat.Size())

	// contents is a []byte
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := w.(http.Flusher)
		if !ok {
			panic("expected http.ResponseWriter to be an http.Flusher")
		}

		connection := &connect.Connection{BufferChannel: make(chan []byte), Buffer: make([]byte, audioSettings.BufferSize)}
		connPool.AddConnection(connection)
		log.Printf("%s has connected to the audio stream\n", r.UserAgent())

		for {
			buf := <-connection.BufferChannel
			if _, err := w.Write(buf); err != nil {
				connPool.DeleteConnection(connection)
				log.Printf("%s's connection to the audio stream has been closed\n", r.UserAgent())
				return
			}
			flusher.Flush() // Triger "chunked" encoding
			// log.Println("emptying buffer")
			clear(connection.Buffer)
		}
	})

	log.Printf("Listening on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func (r *RootCfg) serveHls() {

	port, _ := r.Flags().GetInt("port")
	filepath, _ := r.Flags().GetString("filepath")

	// connPool := connect.NewConnectionPool()

	log.Println("calling go stream...")
	// go connect.Stream(connPool, contents, audioSettings)
	// Array equal to sample rate * 1s

	http.Handle("/", addHeaders(http.FileServer(http.Dir(filepath))))

	slog.Info(fmt.Sprintf("Starting server on port %v\n", port))
	slog.Info(fmt.Sprintf("Serving %s over HTTP on port %v\n", filepath, port))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

}

func addHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	}

}
