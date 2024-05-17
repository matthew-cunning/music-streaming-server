FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

COPY . .

RUN go mod download && go mod verify

RUN GOOS=linux go build -v -o /audio-streamer

EXPOSE 9001

CMD ["/audio-streamer"]

