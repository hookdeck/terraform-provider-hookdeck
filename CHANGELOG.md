## 0.2.0

### BREAKING CHANGES

- rename "archived" to "disabled"

### FEATURES

#### Added new verification support

- Ebay
- Enode
- FrontApp
- Linear
- Orb
- Pylon
- Shopline
- Telnyx
- TokenIo

#### Added JSON source verification support

Useful when Hookdeck supports the verification type but the Terraform provider hasn't been updated yet.

```tf
resource "hookdeck_source_verification" "verification_example" {
  source_id = hookdeck_source.example.id
  verification = {
    json = jsonencode({
      type = "stripe"
      configs = {
        webhook_secret_key = "secret"
      }
    })
  }
}
```

## 0.1.4

### FEATURES

Added verification support for:

- CloudSignal
- Courier
- Favro
- NMI
- Persona
- Repay
- Sanity
- SolidGate
- Square
- Trello
- Twitch
- Wix
