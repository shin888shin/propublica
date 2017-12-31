FROM golang:1.9.2-alpine
ADD propublica /go/bin/propublica
EXPOSE 9090
ENTRYPOINT /go/bin/propublica
# GOOS=linux GOARCH=amd64 go build -o propublica 

# >>> this doesnt work. cant create a working executable with alpine arch
# FROM alpine:3.5
# MAINTAINER Joel Shin
# COPY ./propublica /app/propublica
# RUN chmod +x /app/propublica
# ENV PORT 8080
# EXPOSE 8080
# ENTRYPOINT /app/propublica
