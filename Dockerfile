FROM golang:1.14.3-alpine AS build
WORKDIR /src/
COPY . /src/
RUN unset GOPATH && go build -o /bin/cards .
ENTRYPOINT  ["/bin/cards"]