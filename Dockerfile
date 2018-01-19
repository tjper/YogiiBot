FROM golang:alpine

RUN apk add --update \
	git

# Install package dependencies
RUN go get -u github.com/denisenkom/go-mssqldb

# Copy go packages in container
COPY . /go/src/github.com/penutty/YogiiBot

# Install go packages
RUN go install github.com/penutty/YogiiBot

# Copy twitch bot credentials
COPY ./twitch_pass.txt /go/bin

#
ENTRYPOINT $GOPATH/bin/YogiiBot


