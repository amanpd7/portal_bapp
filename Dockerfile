#builder stage
FROM golang:1.22-alpine AS builder
ENV APPHOME=/app
WORKDIR $APPHOME
COPY . ./
RUN go mod download && go mod verify && go mod tidy
RUN go build -o /main ./main.go

#final stage
FROM alpine:latest
ENV APPHOME=/app
WORKDIR $APPHOME
COPY --from=builder /main ./
COPY ./assets ./assets
COPY ./config.yaml ./config.yaml
RUN chmod 777 ./main
EXPOSE 6969
WORKDIR ${APPHOME}
CMD ["./main"]