# driftlog

Lightweight infrastructure drift detector that compares live cloud resources against Terraform state and outputs a human-readable diff.

---

## Installation

```bash
go install github.com/yourorg/driftlog@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/driftlog.git && cd driftlog && go build -o driftlog .
```

---

## Usage

Point `driftlog` at your Terraform state file and let it compare against your live cloud environment:

```bash
driftlog --state terraform.tfstate --provider aws --region us-east-1
```

**Example output:**

```
[DRIFT DETECTED] aws_instance.web
  ~ instance_type: "t3.micro" → "t3.small"
  ~ tags.Env:      "production" → "staging"

[OK] aws_s3_bucket.assets
[OK] aws_vpc.main

Summary: 1 drifted, 2 in sync
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--state` | Path to Terraform state file | `terraform.tfstate` |
| `--provider` | Cloud provider (`aws`, `gcp`, `azure`) | `aws` |
| `--region` | Target region | `us-east-1` |
| `--output` | Output format (`text`, `json`) | `text` |
| `--quiet` | Only report drifted resources | `false` |

---

## Requirements

- Go 1.21+
- Valid cloud provider credentials (e.g., AWS credentials via `~/.aws/credentials` or environment variables)

---

## License

MIT © 2024 yourorg