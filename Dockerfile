FROM golang:1.19

# Install packages
RUN curl -sL https://deb.nodesource.com/setup_16.x | bash -

RUN apt-get install -y nodejs
RUN apt-get install -y npm || echo "npm already installed"

RUN alias ll="ls -al"

# Copy in source and install deps
RUN mkdir -p /app

COPY ./ /app/
WORKDIR /app

RUN npm install -g serverless@3 && npm install

WORKDIR /app

RUN go get ./...
