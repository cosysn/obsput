# obsput

Upload binaries to Huawei Cloud OBS with CLI tool.

## Installation

```bash
go install
```

## Usage

```bash
# Upload
obsput upload ./bin/app --prefix releases

# List versions
obsput list

# Delete version
obsput delete v1.0.0-abc123-20260212-143000

# Download
obsput download v1.0.0-abc123-20260212-143000

# Configure OBS
obsput obs add --name prod --endpoint "obs.xxx.com" --bucket "bucket" --ak "ak" --sk "sk"
```
