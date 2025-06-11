#!/bin/sh

echo "[Crawler] Starting service..."

echo "[Crawler] Using config at: ./configs/crawler.yaml"

exec ./crawler --config=./configs/crawler.yaml
