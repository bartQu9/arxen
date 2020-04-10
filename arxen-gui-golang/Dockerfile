FROM golang:1.14.0-alpine

# TODO fix dockerfile
LABEL maintainer="Noone"

ENV GO111MODULE=on
ENV PORT=8000

WORKDIR /app

ADD . /app

#RUN go mod init main

ADD go.mod .
ADD go.sum .
RUN go mod download

RUN go build -o main .

EXPOSE 80
EXPOSE 8000

# Command to run the executable
CMD ["./main"]


#CMD ["ping", "1.1.1.1"]