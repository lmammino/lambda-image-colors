FROM golang:1.11.1

# /app will need to be a volume that mounts the entire project path
RUN mkdir -p /app
WORKDIR /app

ENV PATH="/app:${PATH}"

# install simpleproxy to forward external to internal tcp traffic
RUN apt-get update && apt-get install -y simpleproxy

COPY ./entrypoint.sh /opt/entrypoint.sh
RUN chmod +x /opt/entrypoint.sh

ENTRYPOINT ["/opt/entrypoint.sh"]
