FROM golang:1.27-trixie

RUN <<EOF
  apt-get update && apt-get install -y --no-install-recommends nodejs npm
  rm -rf /var/lib/apt/lists/*
  npm install -g aws-cdk
  adduser user
EOF

USER user

WORKDIR /cdk
