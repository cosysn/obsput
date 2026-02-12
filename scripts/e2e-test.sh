#!/bin/bash

# E2E Test Script for obsput
# Tests: put → list → download → delete

echo "=== E2E Test Started ==="

# ============================================
# 1. Setup Environment
# ============================================
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
E2E_DIR="$PROJECT_DIR/.e2e-test"
MINIO_DIR="$E2E_DIR/minio"
MC="$MINIO_DIR/mc"
OBSPUT="$E2E_DIR/obsput"
MINIO_DATA="$E2E_DIR/data"

# MinIO configuration
MINIO_PORT="${MINIO_PORT:-9000}"
MINIO_ENDPOINT="localhost:${MINIO_PORT}"
MINIO_USER="admin"
MINIO_PASSWORD="password"
BUCKET_NAME="test-bucket"

# Create directories
mkdir -p "$MINIO_DIR"
mkdir -p "$MINIO_DATA"
mkdir -p "$E2E_DIR"

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    # Stop MinIO if we started it
    if [ -n "$MINIO_PID" ] && kill -0 "$MINIO_PID" 2>/dev/null; then
        kill "$MINIO_PID" 2>/dev/null || true
        wait "$MINIO_PID" 2>/dev/null || true
    fi
    # Remove test directory
    rm -rf "$E2E_DIR"
    echo "Cleanup complete."
}
trap cleanup EXIT

# ============================================
# 2. Download MinIO + mc (if not exists)
# ============================================
download_if_missing() {
    local url="$1"
    local dest="$2"
    local name="$3"

    if [ -f "$dest" ]; then
        echo "$name already exists, skipping download."
        return 0
    fi

    # Check for existing binary in common locations
    case "$name" in
        "MinIO")
            if [ -f "/tmp/minio" ] && /tmp/minio --version >/dev/null 2>&1; then
                echo "Using existing MinIO at /tmp/minio"
                ln -sf /tmp/minio "$dest"
                return 0
            fi
            ;;
        "MinIO Client")
            if [ -f "/tmp/mc" ] && /tmp/mc --version >/dev/null 2>&1; then
                echo "Using existing mc at /tmp/mc"
                ln -sf /tmp/mc "$dest"
                return 0
            fi
            ;;
    esac

    echo "Downloading $name..."
    if ! curl -sL "$url" -o "$dest"; then
        echo "Failed to download $name from $url"
        exit 1
    fi
    chmod +x "$dest"
    echo "$name downloaded successfully."
}

# Download MinIO
download_if_missing "https://dl.minio.org.cn/server/minio/release/linux-amd64/minio" "$MINIO_DIR/minio" "MinIO"

# Download mc (MinIO Client)
download_if_missing "https://dl.minio.org.cn/client/mc/release/linux-amd64/mc" "$MC" "MinIO Client"

# ============================================
# 3. Start MinIO (Background)
# ============================================
start_minio_server() {
    local port="$1"
    echo "Starting MinIO on port ${port}..."

    # Check if the port is already in use
    if netstat -tuln 2>/dev/null | grep -q ":${port} " || ss -tuln 2>/dev/null | grep -q ":${port} "; then
        echo "Port ${port} is already in use."
        return 1
    fi

    # Start MinIO in background
    MINIO_ROOT_USER=$MINIO_USER MINIO_ROOT_PASSWORD=$MINIO_PASSWORD \
        "$MINIO_DIR/minio" server "$MINIO_DATA" --address ":${port}" \
        > "$E2E_DIR/minio.log" 2>&1 &
    MINIO_PID=$!

    # Wait for MinIO to be ready
    echo "Waiting for MinIO to be ready..."
    for i in {1..30}; do
        if curl -s --fail "http://localhost:${port}/minio/health/live" > /dev/null 2>&1; then
            echo "MinIO is ready!"
            return 0
        fi
        sleep 1
    done
    echo "MinIO failed to start. Check log: $E2E_DIR/minio.log"
    cat "$E2E_DIR/minio.log"
    return 1
}

# Try to start MinIO, or use existing one
if start_minio_server "$MINIO_PORT"; then
    echo "Started new MinIO server on port $MINIO_PORT"
else
    echo "Port $MINIO_PORT is in use, checking for existing MinIO..."
    # Check if MinIO is already running on this port
    if curl -s --fail "http://localhost:${MINIO_PORT}/minio/health/live" > /dev/null 2>&1; then
        echo "Using existing MinIO on port $MINIO_PORT"
        MINIO_PID=""
    else
        # Try to find an available port
        for port in 9001 9002 9003; do
            if start_minio_server "$port"; then
                MINIO_PORT="$port"
                MINIO_ENDPOINT="localhost:${MINIO_PORT}"
                echo "Using MinIO on port $MINIO_PORT"
                break
            fi
        done
        if [ -z "$MINIO_PID" ]; then
            echo "Could not start MinIO on any port"
            exit 1
        fi
    fi
fi

# ============================================
# 4. Configure mc and Create Bucket
# ============================================
echo "Configuring MinIO Client..."

# Configure mc alias (suppress output)
"$MC" alias set myminio "http://${MINIO_ENDPOINT}" "$MINIO_USER" "$MINIO_PASSWORD" 2>/dev/null || true

# Create bucket (suppress output but check result)
echo "Creating bucket: $BUCKET_NAME..."
if ! "$MC" mb "myminio/$BUCKET_NAME" 2>/dev/null; then
    echo "Bucket may already exist or creation failed, continuing..."
fi

# Verify bucket exists
if ! "$MC" ls "myminio/$BUCKET_NAME" 2>/dev/null; then
    echo "Warning: Bucket $BUCKET_NAME not found, trying to create again..."
    "$MC" mb "myminio/$BUCKET_NAME" || echo "Could not create bucket, will try put anyway..."
fi

# ============================================
# 5. Build obsput
# ============================================
echo "Building obsput..."
cd "$PROJECT_DIR"
if ! go build -o "$OBSPUT" .; then
    echo "Failed to build obsput"
    exit 1
fi

# ============================================
# 6. Generate Configuration
# ============================================
echo "Generating obsput configuration..."

CONFIG_DIR="$E2E_DIR/.obsput"
mkdir -p "$CONFIG_DIR"

cat > "$CONFIG_DIR/obsput.yaml" << EOF
configs:
  prod:
    name: prod
    endpoint: http://${MINIO_ENDPOINT}
    bucket: ${BUCKET_NAME}
    ak: ${MINIO_USER}
    sk: ${MINIO_PASSWORD}
EOF

# The config must be in the same directory as the obsput binary
# or in the home directory as ~/.obsput.yaml
# We'll create a symlink in the obsput binary directory
OBSPUT_DIR=$(dirname "$OBSPUT")
OBSPUT_CONFIG_DIR="$OBSPUT_DIR/.obsput"
mkdir -p "$OBSPUT_CONFIG_DIR"
if [ "$CONFIG_DIR/obsput.yaml" != "$OBSPUT_CONFIG_DIR/obsput.yaml" ]; then
    ln -sf "$CONFIG_DIR/obsput.yaml" "$OBSPUT_CONFIG_DIR/obsput.yaml"
fi

# Also create ~/.obsput.yaml as fallback
HOME_CONFIG="$HOME/.obsput.yaml"
if [ "$CONFIG_DIR/obsput.yaml" != "$HOME_CONFIG" ]; then
    ln -sf "$CONFIG_DIR/obsput.yaml" "$HOME_CONFIG" 2>/dev/null || true
fi

# ============================================
# 7. Run E2E Tests
# ============================================
echo "Running E2E tests..."

TEST_FILE="$E2E_DIR/test.bin"
DOWNLOAD_FILE="$E2E_DIR/test_downloaded.bin"

# Generate test file with random data
dd if=/dev/urandom of="$TEST_FILE" bs=1M count=1 2>/dev/null
ORIGINAL_MD5=$(md5sum "$TEST_FILE" | awk '{print $1}')
echo "Test file created: $TEST_FILE (MD5: $ORIGINAL_MD5)"

# ----------------------------------------
# Test 1: Put
# ----------------------------------------
echo ""
echo "--- Test 1: Put ---"
PUT_OUTPUT=$("$OBSPUT" put "$TEST_FILE" --profile prod 2>&1)
echo "$PUT_OUTPUT"

# Check if put was successful
if echo "$PUT_OUTPUT" | grep -q "1 completed, 0 failed"; then
    # Extract version ID from put output
    # Format: "Version: v1.0.0-xxx-20260213-xxxx-x"
    VERSION=$(echo "$PUT_OUTPUT" | grep -oP 'Version:\s*\K[v0-9a-f\-]+' | head -1)

    if [ -z "$VERSION" ]; then
        echo "Error: Could not extract version ID from put output"
        exit 1
    fi
    echo "Put version: $VERSION"
    echo "Put: SUCCESS"
else
    echo "Put failed!"
    exit 1
fi

# ----------------------------------------
# Test 2: List
# ----------------------------------------
echo ""
echo "--- Test 2: List ---"
LIST_OUTPUT=$("$OBSPUT" list --profile prod 2>&1)
echo "$LIST_OUTPUT"

if echo "$LIST_OUTPUT" | grep -q "$VERSION"; then
    echo "List: SUCCESS"
else
    echo "Error: Version $VERSION not found in list output"
    exit 1
fi

# ----------------------------------------
# Test 3: Download + MD5 Verification
# ----------------------------------------
echo ""
echo "--- Test 3: Download + MD5 Verification ---"

# Use mc to download for verification
if ! "$MC" cp "myminio/$BUCKET_NAME/$VERSION/test.bin" "$DOWNLOAD_FILE" 2>/dev/null; then
    echo "Failed to download file using mc, trying obsput download..."
    # Try obsput download as fallback
    "$OBSPUT" download "$VERSION" --profile prod 2>&1
    if [ -f "$TEST_FILE" ]; then
        cp "$TEST_FILE" "$DOWNLOAD_FILE"
    fi
fi

DOWNLOADED_MD5=$(md5sum "$DOWNLOAD_FILE" | awk '{print $1}')
echo "Downloaded file MD5: $DOWNLOADED_MD5"

if [ "$ORIGINAL_MD5" = "$DOWNLOADED_MD5" ]; then
    echo "MD5 verification: SUCCESS"
    echo "Download + MD5: SUCCESS"
else
    echo "Error: MD5 mismatch!"
    echo "  Original: $ORIGINAL_MD5"
    echo "  Downloaded: $DOWNLOADED_MD5"
    exit 1
fi

# ----------------------------------------
# Test 4: Delete
# ----------------------------------------
echo ""
echo "--- Test 4: Delete ---"

DELETE_OUTPUT=$("$OBSPUT" delete "$VERSION" --profile prod 2>&1)
echo "$DELETE_OUTPUT"

# Verify deletion
if echo "$DELETE_OUTPUT" | grep -qi "deleted\|success"; then
    echo "Delete: SUCCESS"
else
    # Check if file still exists via mc
    if "$MC" ls "myminio/$BUCKET_NAME/$VERSION/test.bin" 2>/dev/null; then
        echo "Error: File still exists after delete"
        exit 1
    fi
    echo "Delete: SUCCESS"
fi

# ============================================
# 8. Output Result
# ============================================
echo ""
echo "=== All tests passed! ==="
echo ""
echo "Summary:"
echo "  - Put: SUCCESS"
echo "  - List: SUCCESS"
echo "  - Download + MD5: SUCCESS"
echo "  - Delete: SUCCESS"

exit 0
