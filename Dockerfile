FROM node:22

RUN <<EOF
  curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
  unzip awscliv2.zip
  rm awscliv2.zip
  ./aws/install

  curl -slLo go.tar.gz https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
  tar -C /usr/local -xzf go.tar.gz
  rm go.tar.gz
  ln -s /usr/local/go/bin/go /usr/local/bin/go

  npm install --ignore-scripts --global aws-cdk
EOF

RUN adduser user
USER user

WORKDIR /cdk
