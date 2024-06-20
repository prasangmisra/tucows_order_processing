FROM golang:1.19-alpine

WORKDIR /app

COPY . .

RUN [ ! -f go.mod ] && go mod init payment || echo "go.mod already exists"
RUN go mod tidy
RUN go build -o payment .

EXPOSE 8081

CMD ["./payment"]
