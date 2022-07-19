# Dockerfile 

# steps to build the application in a Docker Image

# --------------------------------------------------------------------------- #

# pull a base image with Go 1.18 already installed for the build step
FROM golang:1.18.4-alpine3.16 as build

# set some environment variables
ENV CGO_ENABLED 0 
ENV GOOS linux

# set the working directory
WORKDIR /src

# copy necessary files into the build container
COPY go.mod ./
COPY go.sum ./
COPY *.go ./

# intall go dependencies
RUN go mod download

# run the tests as part of the build
RUN go test

# build the Go binary
RUN  go build -a -o main .

# standup the final image from scratch
FROM ubuntu as final

RUN apt-get update && apt-get install curl -y

# set the working directory
WORKDIR /bin

# copy over the built image from the build container
COPY --from=build /src/main .

# set the entrypoint for when the container spins up
ENTRYPOINT [ "./main" ]
#EOF