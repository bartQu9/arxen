FROM golang:1.14.0-alpine

# TODO fix dockerfile
LABEL maintainer="arxen_dev"

ENV GO111MODULE=on
ENV PORT=8000

WORKDIR /app

ADD . /app

#RUN go mod init main

ADD go.mod .
ADD go.sum .
RUN go mod download

#RUN go build -o main .

RUN go build -o /app/main.out .

EXPOSE 80
EXPOSE 8000
EXPOSE 7878
EXPOSE 7879

# Command to run the executable
CMD ["./main.out"]
# RUN cd arxen-gui-golang
#CMD ["ls", "/app", "-la"]
