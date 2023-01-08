# Specify the base image for the go app.
FROM golang:alpine
# Specify that we now need to execute any commands in this directory.
WORKDIR /go/src/henrry.online

RUN apk add git

# Copy everything from this project into the filesystem of the container.
COPY . .

RUN go mod tidy

EXPOSE 4000

# Compile the binary exe for our app.
RUN go build -o main .
# Start the application.
CMD ["./main"]