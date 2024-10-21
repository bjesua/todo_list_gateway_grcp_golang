FROM golang:1.19
WORKDIR /app
COPY . .
RUN go mod init todo && go mod tidy
RUN go build -o todo .
CMD [ "./todo" ]
