FROM golang:1.19
WORKDIR /app
COPY . .
RUN go mod init gateway && go mod tidy
RUN go build -o gateway .
CMD [ "./gateway" ]
