#!/bin/bash
set -e

echo "Starting MinIO..."
docker-compose up -d

echo "Waiting for MinIO to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:9000/minio/health/live > /dev/null 2>&1; then
        echo "MinIO is ready!"
        echo "Console: http://localhost:9001 (admin/password)"
        echo "API: http://localhost:9000"
        exit 0
    fi
    sleep 1
done

echo "MinIO failed to start"
exit 1
