FROM node:24-trixie-slim AS node

FROM golang:1.27-trixie

COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules

RUN <<EOF
  ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm
  ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx
  npm install -g aws-cdk@2.1148.1
  adduser user
EOF

USER user

WORKDIR /cdk
