##
## BUILD
##
FROM golang:1.16-alpine AS build

WORKDIR /build

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN go build -o /myfeed cmd/myfeed/main.go

##
## DEPLOY
##
FROM gcr.io/distroless/static:latest AS deploy

WORKDIR /

COPY --from=build /myfeed /myfeed

ENV ADDRESS=":80"

EXPOSE 80

USER nonroot:nonroot

CMD ["/myfeed"]