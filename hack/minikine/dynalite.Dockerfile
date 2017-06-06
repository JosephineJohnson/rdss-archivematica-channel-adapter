FROM node:alpine

ADD . /src

WORKDIR /src

RUN apk add --no-cache git && npm install

ENTRYPOINT ["node", "./dynalite.js"]
