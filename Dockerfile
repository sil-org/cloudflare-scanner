FROM golang:1.19

# Install packages
RUN curl -sL https://deb.nodesource.com/setup_16.x | bash -

RUN apt-get install -y nodejs

RUN alias ll="ls -al"

# Copy in source and install deps
RUN mkdir -p /app

COPY ./ /app/
WORKDIR /app

RUN which npm
RUN npm install -g serverless@3 && npm install

WORKDIR /app

RUN go get ./...
