FROM golang:1.11.1

RUN apt-get update && apt-get install -y git make && \
	rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["make"]

