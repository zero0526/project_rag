#!/bin/sh
echo "Starting Crawler on port ${PORT:-8080}"
exec /app/crawler
