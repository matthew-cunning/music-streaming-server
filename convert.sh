MP3PATH='/Users/matthew/Code/atu/project/project-audio-streamer/music/mp3'   
HLSPATH='/Users/matthew/Code/atu/project/project-audio-streamer/music/hls'   
echo "MP3PATH: $MP3PATH"
echo "HLSPATH: $HLSPATH"

for i in $MP3PATH/*.mp3; 
do 

	BASENAME=$(echo $(basename $i)); 
	NAME=$(echo ${BASENAME%.*})
	echo "NAME: $NAME"
	NOEXTENSION=$(echo "$HLSPATH/$NAME")
	echo "NOEXTENSION: $NOEXTENSION"
	ffmpeg -i "$i" -c:a libmp3lame -b:a 128k -map 0:0 -f segment -segment_time 10 -segment_list "$NOEXTENSION.m3u8" -segment_format mpegts "$NOEXTENSION%03d.ts"
	echo 'END OF LOOP ITERATION';
done
