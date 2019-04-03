FROM debian:latest
ADD test-crawler /app
CMD ["/app", "https://www.semrush.com"]