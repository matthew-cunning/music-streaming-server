package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/matthew-c-atu/project-audio-streamer/internal/pkg/utils"
	"github.com/spf13/cobra"
)

type RootCfg struct{ *cobra.Command }

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "project-audio-streamer",
	Short: "Audio streaming server application which serves audio files over HTTP.",
	Long: `### project-audio-streamer ###
An audio streaming server application which serves audio files over HTTP using HLS.
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
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose logging")
	rootCmd.PersistentFlags().BoolP("debug", "g", false, "Debug logging")
	rootCmd.PersistentFlags().IntP("port", "p", 9001, "The port on which to host the HLS file server")
	rootCmd.PersistentFlags().StringP("dirpath", "d", "music/hls", "The path to the directory being served")
}

const (
	songFilesEndpoint = "/songfiles"
	songNamesEndpoint = "/songnames"
	extension         = ".m3u8"
)

func (r *RootCfg) serveHls() {
	port, err := r.Flags().GetInt("port")
	if err != nil {
		log.Fatal("Couldn't get port flag")
	}

	verbose, _ := r.Flags().GetBool("verbose")
	debug, _ := r.Flags().GetBool("debug")
	dirPath, _ := r.Flags().GetString("dirpath")

	playlistFiles := utils.FindFilesWithExtension(dirPath, extension)

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

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir(dirPath))

	mux.Handle("/", utils.AddHeaders(fs))
	mux.Handle(songFilesEndpoint, utils.AddHeaders(songFilesHandler(marshaledPlaylistFiles)))
	mux.Handle(songNamesEndpoint, utils.AddHeaders(songNamesHandler(marshaledSongNames)))

	if debug {
		r.printFileNames()
	}

	slog.Info(fmt.Sprintf("Starting fileserver on port %v\n", port))

	if verbose {
		slog.Info(fmt.Sprintf("Serving %s over HTTP on port %v\n", dirPath, port))
		slog.Info(fmt.Sprintf("Serving names of files with exentension %v in directory %v over HTTP on port %v%s\n", extension, dirPath, port, songFilesEndpoint))
		slog.Info(fmt.Sprintf("Serving names of songs in directory %v over HTTP on port %v%s\n", dirPath, port, songNamesEndpoint))
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}

func (r *RootCfg) printFileNames() {
	dirPath, _ := r.Flags().GetString("dirpath")
	found := utils.FindFilesWithExtension(dirPath, ".m3u8")
	slog.Info("Found files:")
	for _, v := range found {
		println(v)
	}

	marshaled, err := json.Marshal(found)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Marshaled files:")
	fmt.Printf("%s\n", marshaled)

	var unmarshaled []string

	err = json.Unmarshal(marshaled, &unmarshaled)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Unmarshaled files:")
	for _, v := range unmarshaled {
		println(v)
	}
}

func songFilesHandler(songFiles []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(songFiles)
	}
}

func songNamesHandler(songNames []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(songNames)
	}
}
