FROM golang:1.22.3

WORKDIR /echo
COPY ./paper /echo

RUN go mod tidy

CMD ["go", "run", "/echo/main.go"]