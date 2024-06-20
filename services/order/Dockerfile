FROM golang:1.19-alpine

WORKDIR /app

COPY . .

RUN [ ! -f go.mod ] && go mod init order || echo "go.mod already exists"
RUN go mod tidy
RUN go build -o order .

EXPOSE 8080

CMD ["./order"]
