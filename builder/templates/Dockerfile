FROM golang:1.18-buster as build

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY cmd src/cmd/
COPY models src/models/
COPY restapi src/restapi/

RUN cd src && \
    ls -la restapi && \
    export CGO_LDFLAGS="-static -w -s" && \
    go build -tags osusergo,netgo -o /application cmd/jens-server/main.go; 

FROM ubuntu:21.04

COPY --from=build /application /bin/application

EXPOSE 8080

CMD ["/bin/application"]

