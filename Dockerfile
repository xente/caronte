FROM golang:latest AS build

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

#Use packr2
RUN go get -u github.com/gobuffalo/packr/v2/packr2
RUN packr2 clean
RUN packr2

#Run Build
RUN go build -o main .

FROM golang:latest
WORKDIR /dist
COPY --from=build /build/main .

ENTRYPOINT ["/dist/main"]

