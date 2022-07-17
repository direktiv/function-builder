FROM golang:1.18.2-alpine as build

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY cmd src/cmd/
COPY models src/models/
COPY restapi src/restapi/

RUN cd src && \
    export CGO_LDFLAGS="-static -w -s" && \
    go build -tags osusergo,netgo -o /application cmd/APPNAME-server/main.go; 

FROM ubuntu:21.04

RUN apt-get update && apt-get install ca-certificates -y

# DON'T CHANGE BELOW 
COPY --from=build /application /bin/application

EXPOSE 8080
EXPOSE 9292

CMD ["/bin/application", "--port=8080", "--host=0.0.0.0", "--write-timeout=0"]

