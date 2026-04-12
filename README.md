# envault

> A CLI tool for encrypting and syncing `.env` files across team members using age encryption.

---

## Installation

```bash
go install github.com/yourname/envault@latest
```

Or download a pre-built binary from the [releases page](https://github.com/yourname/envault/releases).

---

## Usage

**Initialize a vault in your project:**
```bash
envault init
```

**Add a recipient (team member's public key):**
```bash
envault add-recipient age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
```

**Encrypt your `.env` file:**
```bash
envault encrypt .env
```

**Decrypt on another machine:**
```bash
envault decrypt .env.age -o .env
```

**Sync encrypted env with your team via git — commit `.env.age`, never `.env`.**

---

## How It Works

envault uses [age](https://github.com/FiloSottile/age) under the hood to encrypt `.env` files with one or more recipients' public keys. Each team member can decrypt using their own private key. No shared secrets, no plaintext credentials in version control.

---

## Requirements

- Go 1.21+
- [age](https://github.com/FiloSottile/age) (bundled, no separate install needed)

---

## License

MIT © [yourname](https://github.com/yourname)