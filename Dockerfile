FROM golang:latest

# Install packages
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get install -y git nodejs netcat
RUN go get -u github.com/golang/dep/cmd/dep

# Copy in source and install deps
RUN mkdir -p /go/src/github.com/silinternational/cloudflare-scanner

COPY ./package.json /go/src/github.com/silinternational/cloudflare-scanner/
WORKDIR /go/src/github.com/silinternational/cloudflare-scanner

RUN npm install -g serverless && npm install

RUN go get github.com/aws/aws-lambda-go/lambda \
           github.com/aws/aws-sdk-go/aws \
           github.com/aws/aws-sdk-go/aws/session \
           github.com/aws/aws-sdk-go/service/ses \
           github.com/cloudflare/cloudflare-go

COPY ./ /go/src/github.com/silinternational/cloudflare-scanner/
