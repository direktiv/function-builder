FROM golang:1.18-buster as build

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY cmd src/cmd/
COPY models src/models/
COPY restapi src/restapi/

RUN cd src && \
    export CGO_LDFLAGS="-static -w -s" && \
    go build -tags osusergo,netgo -o /application cmd/aws-cli-server/main.go; 

# change to aws image
FROM amazon/aws-cli:2.6.1


COPY --from=build /application /bin/application

EXPOSE 8080

# remove image endpoint
ENTRYPOINT []

CMD ["/bin/application", "--port=8080", "--host=0.0.0.0"]

