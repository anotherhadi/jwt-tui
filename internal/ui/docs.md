# JWT Reference

## Structure

A JSON Web Token is three Base64URL-encoded parts joined by dots:

```
header.payload.signature
```

- **Header**: algorithm and token type
- **Payload**: claims (statements about an entity)
- **Signature**: HMAC or RSA/ECDSA over header + payload

The header and payload are readable by anyone. JWTs are _signed_, not _encrypted_.
Use JWE if you need confidentiality.

## Header

```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

Common header parameters:

| Param | Description                                |
| ----- | ------------------------------------------ |
| `alg` | Signing algorithm (`HS256`, `RS256`, etc.) |
| `typ` | Token type: always `JWT`                   |
| `kid` | Key ID: hint for which key to use          |
| `cty` | Content type: used for nested JWTs         |

## Payload (Claims)

**Registered claims** (all optional, but recommended):

| Claim | Type   | Description                             |
| ----- | ------ | --------------------------------------- |
| `iss` | string | Issuer: who created the token           |
| `sub` | string | Subject: principal the token is about   |
| `aud` | string | Audience: intended recipient(s)         |
| `exp` | number | Expiration time (Unix timestamp)        |
| `nbf` | number | Not before: token valid after this time |
| `iat` | number | Issued at (Unix timestamp)              |
| `jti` | string | JWT ID: unique identifier               |

**Private claims** are any additional fields agreed upon by the parties.

## Algorithms

| Algorithm | Type           | Key type                  |
| --------- | -------------- | ------------------------- |
| `HS256`   | HMAC + SHA-256 | Shared secret             |
| `HS384`   | HMAC + SHA-384 | Shared secret             |
| `HS512`   | HMAC + SHA-512 | Shared secret             |
| `RS256`   | RSA + SHA-256  | RSA key pair              |
| `RS384`   | RSA + SHA-384  | RSA key pair              |
| `RS512`   | RSA + SHA-512  | RSA key pair              |
| `ES256`   | ECDSA + P-256  | EC key pair               |
| `ES384`   | ECDSA + P-384  | EC key pair               |
| `ES512`   | ECDSA + P-521  | EC key pair               |
| `none`    | No signature   | ⚠ Never use in production |

> This tool supports **HS256**, **HS384**, and **HS512**.

## Signature Computation

For HMAC algorithms:

```
signature = HMAC-SHA256(
  base64url(header) + "." + base64url(payload),
  secret
)
```

The final token:

```
base64url(header) + "." + base64url(payload) + "." + base64url(signature)
```

## Security

- **Never use `alg: none`**: disables signature verification entirely.
- Use **long, random secrets**: at least 256 bits (32 bytes) for HS256.
- Always validate **`exp`** (expiration) and **`nbf`** (not before).
- Validate **`iss`** and **`aud`** to prevent token reuse across services.
- The payload is **base64-encoded, not encrypted**: never store passwords or PII.
- Prefer **asymmetric algorithms** (RS256, ES256) for public-facing APIs.
- Store secrets in environment variables or a secrets manager, never in code.

## Brute-forcing a JWT Secret

If a token is signed with a weak HMAC secret, it can be recovered offline.
Both **hashcat** and **john** accept the raw JWT string as input:

```bash
# hashcat mode 16500 targets JWT (HS256/384/512)
hashcat -a 0 -m 16500 <token> wordlist.txt

# john the ripper
john --format=HMAC-SHA256 --wordlist=wordlist.txt jwt.txt
```

This only works against **HS\*** algorithms where the secret is a simple password or passphrase.

## Configuration

jwt-tui looks for a config file at `~/.config/jwt-tui/config.yaml`.
If the file does not exist the built-in defaults are used automatically.

To get a starting point you can edit, run:

```
jwt-tui --add-default-config
```

This writes the default config to `~/.config/jwt-tui/config.yaml` (or to the path given with `--config`).
You can then open that file in any text editor and change the values you want.
