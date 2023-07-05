FROM golang:latest as backend

WORKDIR /usr/src/adhoc

COPY ./backend .

RUN CGO_ENABLED=0 go build


FROM node:latest as frontend

WORKDIR /usr/src/adhoc

COPY ./frontend .

RUN  npm i && npm run build



FROM alpine:3.14 as host

COPY --from=backend /usr/src/adhoc/server /usr/local/bin/server

COPY --from=frontend /usr/src/adhoc/build /usr/local/bin/build

WORKDIR /usr/local/bin

RUN apk update && apk add ffmpeg && rm -rf /var/cache/apk/*


CMD ["server"]
