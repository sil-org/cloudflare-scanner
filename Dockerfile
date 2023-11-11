FROM golang:1.19

RUN curl -o- -L https://slss.io/install | VERSION=3.36.0 bash && \
  mv $HOME/.serverless/bin/serverless /usr/local/bin && \
  ln -s /usr/local/bin/serverless /usr/local/bin/sls

WORKDIR /app
COPY ./ .
RUN go get ./...
