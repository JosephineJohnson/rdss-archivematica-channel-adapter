FROM node:alpine

ADD . /src

WORKDIR /src

RUN npm install

ENTRYPOINT ["node", "./dynalite.js"]
