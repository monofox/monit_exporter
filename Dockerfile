FROM golang:alpine

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o monit_exporter .

WORKDIR /dist

RUN cp /build/monit_exporter .

EXPOSE 9388

ENTRYPOINT ["/dist/monit_exporter"]