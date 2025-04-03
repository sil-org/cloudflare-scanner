FROM node

RUN <<EOF
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
    unzip awscliv2.zip
    ./aws/install
    npm install -g aws-cdk typescript
EOF

RUN adduser user
USER user

WORKDIR /cdk
