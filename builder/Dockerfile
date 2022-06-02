FROM ubuntu:21.10

RUN apt-get update && apt-get install openjdk-18-jre wget jq curl golang-1.17 -y

# install karate    
RUN wget https://github.com/karatelabs/karate/releases/download/v1.2.0/karate-1.2.0.jar

# install go-swagger
RUN download_url=$(curl -s https://api.github.com/repos/go-swagger/go-swagger/releases/latest | \
  jq -r '.assets[] | select(.name | contains("'"$(uname | tr '[:upper:]' '[:lower:]')"'_amd64")) | .browser_download_url') && \
    curl -o /usr/local/bin/swagger -L'#' "$download_url" && \
    chmod +x /usr/local/bin/swagger

COPY builder.sh /
RUN chmod 755 builder.sh 

RUN mkdir /go /.cache
ENV GOPATH=/go

RUN chmod 777 /go
RUN chmod 777 /.cache/

COPY templates /templates 



ENTRYPOINT ["/builder.sh"]