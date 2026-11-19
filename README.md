> **Deprecated** — this repository has been superseded. The official Go SDK now lives at
> **[github.com/senviok/senviok](https://github.com/senviok/senviok)** and is released as
> `v0.1.0`. Install it with `go get github.com/senviok/senviok`. This repo is kept for
> historical reference only and will not receive updates.

# Senviok Go SDK

Official Go client for the [Senviok](https://senviok.live) messaging API —
email, SMS and WhatsApp, plus the resources around them. It mirrors the official
JavaScript and Python SDKs' public facade 1:1, so all three stay in lockstep.

- Uses **only the Go standard library** — no external dependencies
- All 11 public resources exposed as client fields
- Automatic snake→camelCase wire format handled for you
- Webhook `HMAC-SHA256` signature verification
- Typed errors (`*AuthenticationError`, `*RateLimitError`, `*APIError`) for
  every non-2xx response

## Installation

```bash
go get github.com/senviok/senviok
```

## Quick start

```go
package main

import (
    "context"
    "fmt"

    "github.com/Chekwube-Manuel/senviok-go-sdk/senviok"
)

func main() {
    client := senviok.New("svk_live_xxx")

    // Email
    resp, err := client.Emails.Send(context.Background(), senviok.EmailOptions{
        From:    "Acme <onboarding@acme.com>",
        To:      "dev@acme.com", // or []string{"dev@acme.com", "ops@acme.com"}
        Subject: "Hello from Senviok",
        HTML:    "<p>Hello!</p>",
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.ID) // msg_...

    // SMS
    if _, err := client.SMS.Send(context.Background(),
        "+2348012345678", "Senviok", "Your code is 123456"); err != nil {
        panic(err)
    }

    // WhatsApp
    if _, err := client.WhatsApp.Send(context.Background(),
        "+2348012345678", "Senviok", "Hello!"); err != nil {
        panic(err)
    }
}
```

## Resources

| Resource | Methods |
| --- | --- |
| `Emails` | `Send(ctx, EmailOptions)` |
| `SMS` | `Send(ctx, to, from, text)` |
| `WhatsApp` | `Send(ctx, to, from, text)` |
| `Messages` | `List(ctx, MessageListParams)` |
| `Webhooks` | `Create(ctx, url, events)`, `List(ctx)`, `Delete(ctx, id)`, `DeliveryLogs(ctx, webhookID)`, `VerifySignature(body, signature, secret)` |
| `Templates` | `Create(ctx, name, subject, htmlContent)`, `List(ctx)`, `Get(ctx, id)`, `Update(ctx, id, UpdateTemplateRequest)`, `Delete(ctx, id)` |
| `Domains` | `Create(ctx, name)`, `List(ctx)`, `GetDKIM(ctx, id)`, `Verify(ctx, id)` |
| `Audiences` | `Create(ctx, name)`, `List(ctx)`, `Delete(ctx, id)` |
| `Contacts` | `Create(ctx, audienceID, CreateContactRequest)`, `List(ctx, audienceID)`, `Delete(ctx, audienceID, id)` |
| `Suppressions` | `Create(ctx, email, reason)`, `List(ctx)`, `Delete(ctx, id)` |
| `APIKeys` | `Create(ctx, name)`, `List(ctx)`, `Delete(ctx, id)` |

### Email options

`Emails.Send` accepts the same options as the JS/Python SDKs:

- `From`, `FromName`, `To`, `Subject`, `HTML`, `Text`
- `CC`, `BCC`, `ReplyTo` — each accepts a `string` or `[]string`
- `TemplateID`, `TemplateData`
- `AddUnsubscribeFooter`, `AddListUnsubscribeHeader` (pointers)
- `Attachments` (`[]EmailAttachment`) and `AttachmentIDs` (IDs returned by the upload endpoint)

Inline attachments use base64 content or a public HTTPS URL; set exactly one:

```go
attachment := senviok.NewEmailAttachment("invoice.pdf", pdfBytes, "application/pdf")
_, err := client.Emails.Send(ctx, senviok.EmailOptions{
    From: "billing@example.com", To: "customer@example.com",
    Subject: "Your invoice", Attachments: []senviok.EmailAttachment{attachment},
})
```

For larger files, upload each file as multipart field `file` to `POST /v1/emails/attachments`, then send the returned ID in `AttachmentIDs`. The API allows up to 10 files and 10 MB total per email.

```go
uploaded, err := client.Emails.UploadAttachment(ctx, "invoice.pdf", pdfBytes, "application/pdf")
if err != nil { return err }
_, err = client.Emails.Send(ctx, senviok.EmailOptions{
    From: "billing@example.com", To: "customer@example.com",
    Subject: "Your invoice", AttachmentIDs: []string{uploaded.ID},
})
```

### Webhook signature verification

```go
func handler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    signature := r.Header.Get("X-Senviok-Signature")
    secret := "whsec-your-webhook-secret"

    if senviok.VerifyWebhookSignature(body, signature, secret) {
        // signature is valid — process the event
    }
}
```

## Local development

Point the client at a locally running API:

```go
client := senviok.New("dev-key", senviok.WithBaseURL("http://localhost:5033"))
```

## Options

| Option | Purpose |
| --- | --- |
| `WithBaseURL(url)` | Override the API base URL (default `https://api.senviok.live`) |
| `WithHTTPClient(hc)` | Provide a custom `*http.Client` (timeout, transport) |

## Errors

| Type | When |
| --- | --- |
| `*AuthenticationError` | Missing/invalid API key (HTTP 401/403) |
| `*RateLimitError` | Tenant rate limit exceeded (HTTP 429, 100 req/min) |
| `*APIError` | Any other non-2xx response (carries `Code` and raw `Body`) |

All satisfy the `senviok.Error` interface (`StatusCode() int`). Use `errors.As`
to inspect a specific type.

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
