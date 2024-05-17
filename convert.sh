function print_usage() {
	echo "Usage: convert.sh {mp3_input_directory} {hls_output_directory}
convert.sh -h || --help : Print detailed help"
}

function print_help() {
	echo "### convert.sh ###

Author: Matthew Cunningham

Usage: convert.sh {mp3_input_directory} {hls_output_directory}

Example: ./convert.sh ./music/mp3 ./music/hls

If {hls_output_directory} doesn't exist, it will be created.

If an .m3u8 playlist file of a found mp3 in already exists in {hls_output_directory}, generation of that playlist file will be skipped."
}

function generate_playlists() {

	if [ ! -d "$2" ]; then
		mkdir -p $2
		echo "Created output directory $(realpath $2)"
	fi

	MP3PATH=$(realpath $1)
	HLSPATH=$(realpath $2)

	echo "MP3PATH: $MP3PATH"
	echo "HLSPATH: $HLSPATH"


	if [ ! -d "$MP3PATH" ]; then
		echo "Error: Input filepath $MP3PATH does not exist."
		return 1
	fi

	if [ ! -d "$HLSPATH" ]; then
		echo "Error: Output filepath $HLSPATH does not exist."
		return 1
	fi

	for i in $MP3PATH/*.mp3; 
	do 
		BASENAME=$(echo $(basename $i)); 
		FILENAME=$(echo ${BASENAME%.*})
		echo "FILENAME: $FILENAME"
		FILENAME_NOEXTENSION=$(echo "$HLSPATH/$FILENAME")
		echo "FILENAME_NOEXTENSION: $FILENAME_NOEXTENSION"

		FILENAME_M3U8=$(echo "$FILENAME.m3u8")
		echo "FILENAME_M3U8: $FILENAME_M3U8"
		if [ -a "$HLSPATH/$FILENAME_M3U8" ]; then 
			echo "Info: Skipping generation of playlist file"
			echo "Already exists: $HLSPATH/$FILENAME_M3U8\n"
			continue
		fi
		echo "Generating playist file for $i ..."
		ffmpeg -i "$i" -c:a libmp3lame -b:a 128k -map 0:0 -f segment -segment_time 10 -segment_list "$FILENAME_NOEXTENSION.m3u8" -segment_format mpegts "$FILENAME_NOEXTENSION%03d.ts"
	done
}

if [[ $# -eq 1 ]]; then
	if [ "$1" == "-h" ] ||  [ "$1" == "--help" ]; then
		print_help
		exit 0
	fi
fi

if [[ $# -ne 2 ]]; then
	print_usage
	exit 1
fi

generate_playlists $1 $2
