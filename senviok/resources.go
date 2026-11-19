package senviok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// decodeList normalizes the inconsistent list payloads ("data", "logs" or a
// bare array) into a slice of v by unmarshaling each element.
func decodeList[T any](payload []byte) ([]T, error) {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	var list []any
	switch r := raw.(type) {
	case []any:
		list = r
	case map[string]any:
		for _, key := range []string{"data", "logs"} {
			if v, ok := r[key].([]any); ok {
				list = v
				break
			}
		}
	default:
		return nil, nil
	}
	out := make([]T, 0, len(list))
	for _, item := range list {
		b, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		var v T
		if err := json.Unmarshal(b, &v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// get is a convenience wrapper for caller that only reads a typed single
// response body.
func unmarshal[T any](raw []byte) (*T, error) {
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// ── Emails ────────────────────────────────────────────────────────────────

type EmailsResource struct{ client *Client }

// Send sends an email with the given options. To/CC/BCC/ReplyTo accept a string
// or []string.
func (r *EmailsResource) Send(ctx context.Context, opts EmailOptions) (*EmailResponse, error) {
	raw, err := r.client.Do(ctx, "POST", "/emails", nil, opts)
	if err != nil {
		return nil, err
	}
	return unmarshal[EmailResponse](raw)
}

// UploadAttachment uploads one file; pass the returned ID through AttachmentIDs when sending.
func (r *EmailsResource) UploadAttachment(ctx context.Context, filename string, content []byte, contentType string) (*UploadedEmailAttachment, error) {
	raw, err := r.client.DoMultipart(ctx, "/emails/attachments", filename, contentType, content)
	if err != nil {
		return nil, err
	}
	return unmarshal[UploadedEmailAttachment](raw)
}

// ── SMS ───────────────────────────────────────────────────────────────────

type SMSResource struct{ client *Client }

func (r *SMSResource) Send(ctx context.Context, to, from, text string) (*SmsResponse, error) {
	raw, err := r.client.Do(ctx, "POST", "/sms", nil, SMSRequest{To: to, From: from, Text: text})
	if err != nil {
		return nil, err
	}
	return unmarshal[SmsResponse](raw)
}

// ── WhatsApp ──────────────────────────────────────────────────────────────

type WhatsAppResource struct{ client *Client }

func (r *WhatsAppResource) Send(ctx context.Context, to, from, text string) (*SmsResponse, error) {
	raw, err := r.client.Do(ctx, "POST", "/whatsapp", nil, SMSRequest{To: to, From: from, Text: text})
	if err != nil {
		return nil, err
	}
	return unmarshal[SmsResponse](raw)
}

// ── Messages ──────────────────────────────────────────────────────────────

type MessagesResource struct{ client *Client }

// List returns the message/analytics logs. Zero-valued params use the same
// defaults as the Python SDK (skip=0, take=20, sortOrder=desc).
func (r *MessagesResource) List(ctx context.Context, params MessageListParams) ([]MessageLog, error) {
	q := url.Values{}
	skip := params.Skip
	if skip == 0 {
		skip = 0
	}
	take := params.Take
	if take == 0 {
		take = 20
	}
	sortOrder := params.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	q.Set("skip", fmt.Sprintf("%d", skip))
	q.Set("take", fmt.Sprintf("%d", take))
	q.Set("sortOrder", sortOrder)
	add := func(k, v string) {
		if v != "" {
			q.Set(k, v)
		}
	}
	add("channel", params.Channel)
	add("status", params.Status)
	add("toAddress", params.ToAddress)
	add("fromAddress", params.FromAddress)
	add("subject", params.Subject)
	add("startDate", params.StartDate)
	add("endDate", params.EndDate)

	raw, err := r.client.Do(ctx, "GET", "/analytics/logs", q, nil)
	if err != nil {
		return nil, err
	}
	return decodeList[MessageLog](raw)
}

// MessageListParams filters for Messages.List. Empty values are omitted to use
// server defaults.
type MessageListParams struct {
	Skip        int
	Take        int    // default 20 when 0
	SortOrder   string // default "desc" when empty
	Channel     string // e.g. "email", "sms", "whatsapp"
	Status      string
	ToAddress   string
	FromAddress string
	Subject     string
	StartDate   string
	EndDate     string
}

// ── Webhooks ──────────────────────────────────────────────────────────────

type WebhooksResource struct{ client *Client }

func (r *WebhooksResource) Create(ctx context.Context, webURL string, events []string) (*Webhook, error) {
	raw, err := r.client.Do(ctx, "POST", "/webhooks", nil, CreateWebhookRequest{URL: webURL, Events: events})
	if err != nil {
		return nil, err
	}
	return unmarshal[Webhook](raw)
}

func (r *WebhooksResource) List(ctx context.Context) ([]Webhook, error) {
	raw, err := r.client.Do(ctx, "GET", "/webhooks", nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeList[Webhook](raw)
}

func (r *WebhooksResource) Delete(ctx context.Context, id string) error {
	_, err := r.client.Do(ctx, "DELETE", "/webhooks/"+id, nil, nil)
	return err
}

func (r *WebhooksResource) DeliveryLogs(ctx context.Context, webhookID string) ([]WebhookDelivery, error) {
	raw, err := r.client.Do(ctx, "GET", "/webhooks/"+webhookID+"/deliveries", nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeList[WebhookDelivery](raw)
}

// VerifySignature is an alias for the package-level constant-time HMAC helper.
func (r *WebhooksResource) VerifySignature(body []byte, signature, secret string) bool {
	return VerifyWebhookSignature(body, signature, secret)
}

// ── Templates ─────────────────────────────────────────────────────────────

type TemplatesResource struct{ client *Client }

func (r *TemplatesResource) Create(ctx context.Context, name, subject, htmlContent string) (*Template, error) {
	raw, err := r.client.Do(ctx, "POST", "/templates", nil, CreateTemplateRequest{
		Name: name, Subject: subject, HTMLContent: htmlContent,
	})
	if err != nil {
		return nil, err
	}
	return unmarshal[Template](raw)
}

func (r *TemplatesResource) List(ctx context.Context) ([]Template, error) {
	raw, err := r.client.Do(ctx, "GET", "/templates", nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeList[Template](raw)
}

func (r *TemplatesResource) Get(ctx context.Context, id string) (*Template, error) {
	raw, err := r.client.Do(ctx, "GET", "/templates/"+id, nil, nil)
	if err != nil {
		return nil, err
	}
	return unmarshal[Template](raw)
}

func (r *TemplatesResource) Update(ctx context.Context, id string, req UpdateTemplateRequest) (*Template, error) {
	raw, err := r.client.Do(ctx, "PUT", "/templates/"+id, nil, req)
	if err != nil {
		return nil, err
	}
	return unmarshal[Template](raw)
}

func (r *TemplatesResource) Delete(ctx context.Context, id string) error {
	_, err := r.client.Do(ctx, "DELETE", "/templates/"+id, nil, nil)
	return err
}

// ── Domains ───────────────────────────────────────────────────────────────

type DomainsResource struct{ client *Client }

func (r *DomainsResource) Create(ctx context.Context, name string) (*Domain, error) {
	raw, err := r.client.Do(ctx, "POST", "/domains", nil, CreateDomainRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return unmarshal[Domain](raw)
}

func (r *DomainsResource) List(ctx context.Context) ([]Domain, error) {
	raw, err := r.client.Do(ctx, "GET", "/domains", nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeList[Domain](raw)
}

func (r *DomainsResource) GetDKIM(ctx context.Context, id string) (*DkimResponse, error) {
	raw, err := r.client.Do(ctx, "GET", "/domains/"+id+"/dkim", nil, nil)
	if err != nil {
		return nil, err
	}
	return unmarshal[DkimResponse](raw)
}

func (r *DomainsResource) Verify(ctx context.Context, id string) (*VerifyDomainResponse, error) {
	raw, err := r.client.Do(ctx, "POST", "/domains/"+id+"/verify", nil, nil)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		raw = []byte("{}")
	}
	return unmarshal[VerifyDomainResponse](raw)
}

// ── Audiences ─────────────────────────────────────────────────────────────

type AudiencesResource struct{ client *Client }

func (r *AudiencesResource) Create(ctx context.Context, name string) (*Audience, error) {
	raw, err := r.client.Do(ctx, "POST", "/audiences", nil, CreateAudienceRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return unmarshal[Audience](raw)
}

func (r *AudiencesResource) List(ctx context.Context) ([]Audience, error) {
	raw, err := r.client.Do(ctx, "GET", "/audiences", nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeList[Audience](raw)
}

func (r *AudiencesResource) Delete(ctx context.Context, id string) error {
	_, err := r.client.Do(ctx, "DELETE", "/audiences/"+id, nil, nil)
	return err
}

// ── Contacts ──────────────────────────────────────────────────────────────

type ContactsResource struct{ client *Client }

func (r *ContactsResource) Create(ctx context.Context, audienceID string, req CreateContactRequest) (*Contact, error) {
	raw, err := r.client.Do(ctx, "POST", "/audiences/"+audienceID+"/contacts", nil, req)
	if err != nil {
		return nil, err
	}
	return unmarshal[Contact](raw)
}

func (r *ContactsResource) List(ctx context.Context, audienceID string) ([]Contact, error) {
	raw, err := r.client.Do(ctx, "GET", "/audiences/"+audienceID+"/contacts", nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeList[Contact](raw)
}

func (r *ContactsResource) Delete(ctx context.Context, audienceID, id string) error {
	_, err := r.client.Do(ctx, "DELETE", "/audiences/"+audienceID+"/contacts/"+id, nil, nil)
	return err
}

// ── Suppressions ──────────────────────────────────────────────────────────

type SuppressionsResource struct{ client *Client }

func (r *SuppressionsResource) Create(ctx context.Context, email, reason string) (*Suppression, error) {
	raw, err := r.client.Do(ctx, "POST", "/suppressions", nil, CreateSuppressionRequest{Email: email, Reason: reason})
	if err != nil {
		return nil, err
	}
	return unmarshal[Suppression](raw)
}

func (r *SuppressionsResource) List(ctx context.Context) ([]Suppression, error) {
	raw, err := r.client.Do(ctx, "GET", "/suppressions", nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeList[Suppression](raw)
}

func (r *SuppressionsResource) Delete(ctx context.Context, id string) error {
	_, err := r.client.Do(ctx, "DELETE", "/suppressions/"+id, nil, nil)
	return err
}

// ── API keys ──────────────────────────────────────────────────────────────

type APIKeysResource struct{ client *Client }

func (r *APIKeysResource) Create(ctx context.Context, name string) (*APIKeyCreated, error) {
	raw, err := r.client.Do(ctx, "POST", "/api-keys", nil, CreateAPIKeyRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return unmarshal[APIKeyCreated](raw)
}

func (r *APIKeysResource) List(ctx context.Context) ([]APIKey, error) {
	raw, err := r.client.Do(ctx, "GET", "/api-keys", nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeList[APIKey](raw)
}

func (r *APIKeysResource) Delete(ctx context.Context, id string) error {
	_, err := r.client.Do(ctx, "DELETE", "/api-keys/"+id, nil, nil)
	return err
}
