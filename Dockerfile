FROM golang:1.18 as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy over the swagger specs
RUN mkdir public && \
	cp $GOPATH/pkg/mod/go.buf.build/jonwhitty/go-grpc-gateway/corux/gps-tracker-auth@*/auth/v1/auth.swagger.json public/ && \
	cp $GOPATH/pkg/mod/go.buf.build/jonwhitty/go-grpc-gateway/corux/gps-tracker-position@*/position/v1/position.swagger.json public/

COPY main.go ./

# Create a staticly linked binary
ENV CGO_ENABLED=0
RUN go build -o server main.go

FROM alpine as run

WORKDIR /app

COPY --from=build /app/public /app/public
COPY --from=build /app/server /app/server

ENTRYPOINT [ "/app/server" ]