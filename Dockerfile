FROM alpine:latest
ADD web-crawler /app
CMD ["/app", "https://www.habr.com", "100"]