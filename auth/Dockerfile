FROM golang:1.19
WORKDIR /app
COPY . .
RUN go mod init auth && go mod tidy
RUN go build -o auth .
CMD [ "./auth" ]
