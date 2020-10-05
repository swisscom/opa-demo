FROM golang:alpine3.12 as builder
RUN mkdir /app 
ADD . /app/
WORKDIR /app 
RUN go build -o main .
RUN chmod 755 main

FROM alpine
COPY --from=builder /app/main /usr/bin/opa-api-example
ENTRYPOINT [ "/usr/bin/opa-api-example" ]
