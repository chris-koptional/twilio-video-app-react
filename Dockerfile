FROM golang:latest as backend

WORKDIR /usr/src/adhoc

COPY ./backend .

RUN CGO_ENABLED=0 go build


FROM node:latest as frontend

WORKDIR /usr/src/adhoc

COPY ./frontend .

RUN  npm i && npm run build



FROM gcr.io/distroless/cc-debian10 as host

COPY --from=backend /usr/src/adhoc/server /usr/local/bin/server

COPY --from=frontend /usr/src/adhoc/build /usr/local/bin/build

WORKDIR /usr/local/bin

CMD ["server"]
