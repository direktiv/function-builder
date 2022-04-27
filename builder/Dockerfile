FROM ubuntu:21.10

RUN apt-get update && apt-get install jq curl golang-1.17 -y

# install go-swagger
RUN download_url=$(curl -s https://api.github.com/repos/go-swagger/go-swagger/releases/latest | \
  jq -r '.assets[] | select(.name | contains("'"$(uname | tr '[:upper:]' '[:lower:]')"'_amd64")) | .browser_download_url') && \
    curl -o /usr/local/bin/swagger -L'#' "$download_url" && \
    chmod +x /usr/local/bin/swagger

COPY builder.sh /
RUN chmod 755 builder.sh 

RUN mkdir /go
ENV GOPATH=/go

COPY templates /templates 


ENTRYPOINT ["/builder.sh"]