FROM golang:1.20.2
WORKDIR /app
COPY . /app/
WORKDIR /app/
RUN apt-get update -y
RUN apt-get install nmap -y
RUN apt-get install iputils-ping -y
RUN apt-get install dnsutils -y
RUN export GODEBUG=netdns=go
RUN go mod download
RUN go build -buildvcs=false -o main .
EXPOSE 80
EXPOSE 443
CMD ["/app/main"]

# docker build -t ipmonitoring:latest . --no-cache