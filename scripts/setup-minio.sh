#!/bin/bash

# Wait for MinIO to be ready
./scripts/start-minio.sh

# Create test bucket
mc alias set myminio http://localhost:9000 admin password || true
mc mb myminio/test-bucket --ignore-existing

echo "Test bucket created!"
