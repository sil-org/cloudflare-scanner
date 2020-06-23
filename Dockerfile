FROM golang:latest

ENV GO111MODULE on

# Install packages
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get install -y git nodejs netcat
#RUN go get -u github.com/golang/dep/cmd/dep

RUN alias ll="ls -al"

# Copy in source and install deps
RUN mkdir -p /app

COPY ./ /app/
WORKDIR /app

RUN npm install -g serverless && npm install

WORKDIR /app

RUN go get ./...
