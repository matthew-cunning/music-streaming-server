/*
Copyright © 2024 NME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
		// root.printFileNames()

		root.serveHls()
		// root.serveFileNames()
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
	rootCmd.PersistentFlags().IntP("hlsport", "p", 9001, "The port on which to host the HLS file server")
	rootCmd.PersistentFlags().IntP("filenameport", "n", 9002, "The port on which to host the filename server")
	rootCmd.PersistentFlags().StringP("filepath", "f", "music/hls", "path to the audio file")
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
	port, _ := r.Flags().GetInt("hlsport")
	filepath, _ := r.Flags().GetString("filepath")

	// connPool := connect.NewConnectionPool()

	// log.Println("calling go stream...")
	// go connect.Stream(connPool, contents, audioSettings)
	// Array equal to sample rate * 1s
	extension := ".m3u8"
	dirPath, _ := r.Flags().GetString("filepath")
	playlistFiles := findFilesWithExtension(dirPath, extension)

	var songNames []string
	for f := range playlistFiles {
		songNames = append(songNames, strings.TrimSuffix(playlistFiles[f], extension))

	}

	marshaledPlaylistFiles, err := json.Marshal(playlistFiles)
	if err != nil {
		log.Fatal(err)
	}

	marshaledSongNames, err := json.Marshal(songNames)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", addHeaders(http.FileServer(http.Dir(filepath))))

	songFilesEndpoint := "/songfiles"
	songNamesEndpoint := "/songnames"
	http.HandleFunc(songFilesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Write(marshaledPlaylistFiles)
	})

	http.HandleFunc(songNamesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Write(marshaledSongNames)
	})
	slog.Info(fmt.Sprintf("Starting fileserver on port %v\n", port))
	slog.Info(fmt.Sprintf("Serving %s over HTTP on port %v\n", filepath, port))
	slog.Info(fmt.Sprintf("Serving names of files with exentension %v in directory %v over HTTP on port %v%s\n", extension, dirPath, port, songFilesEndpoint))
	slog.Info(fmt.Sprintf("Serving names of songs in directory %v over HTTP on port %v%s\n", dirPath, port, songNamesEndpoint))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func (r *RootCfg) serveFileNames() {
	extension := ".m3u8"
	port, _ := r.Flags().GetInt("filenameport")
	dirPath, _ := r.Flags().GetString("filepath")
	playlistFiles := findFilesWithExtension(dirPath, extension)
	marshaled, err := json.Marshal(playlistFiles)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		w.Write(marshaled)
	})

	slog.Info(fmt.Sprintf("Starting fileserver on port %v\n", port))
	slog.Info(fmt.Sprintf("Serving names of files with exentension %v in directory %v over HTTP on port %v\n", extension, dirPath, port))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

}
func (r *RootCfg) printFileNames() {
	dirPath, _ := r.Flags().GetString("filepath")

	// files, err := filepath.Glob(fmt.Sprintf("%v/*.m3u8", dirPath))
	// files, err := os.ReadDir(dirPath)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// for _, f := range files {
	// 	fmt.Println(f)
	// }
	//
	found := findFilesWithExtension(dirPath, ".m3u8")
	for f := range found {
		fmt.Println(found[f])
	}

	marshaled, err := json.Marshal(found)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", marshaled)

	var unmarshaled []string
	err = json.Unmarshal(marshaled, &unmarshaled)
	if err != nil {
		log.Fatal(err)
	}
	for f := range unmarshaled {
		println(unmarshaled[f])
	}
}

func findFilesWithExtension(root, extension string) []string {
	var found []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == extension {
			found = append(found, d.Name())
		}
		return nil
	})
	return found
}

func addHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	}
}
