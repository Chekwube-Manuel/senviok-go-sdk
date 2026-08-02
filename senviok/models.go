package senviok

// This file mirrors the Python SDK's models.py. Request structs serialize to
// lowercase camelCase keys (the SDK's to_camel convention); response structs
// read the API's camelCase JSON fields.

// ─── Requests ─────────────────────────────────────────────────────────────

// EmailOptions mirrors the Python EmailOptions. To/CC/BCC/ReplyTo may hold a
// single string or []string (the API's `object` fields). HTML maps to the
// request key "html"; Text to "text".
type EmailOptions struct {
	From                     string            `json:"from"`
	FromName                 string            `json:"fromName,omitempty"`
	To                       any               `json:"to,omitempty"`
	Subject                  string            `json:"subject,omitempty"`
	HTML                     string            `json:"html,omitempty"`
	Text                     string            `json:"text,omitempty"`
	CC                       any               `json:"cc,omitempty"`
	BCC                      any               `json:"bcc,omitempty"`
	ReplyTo                  any               `json:"replyTo,omitempty"`
	TemplateID               string            `json:"templateId,omitempty"`
	TemplateData             map[string]string `json:"templateData,omitempty"`
	AddUnsubscribeFooter     *bool             `json:"addUnsubscribeFooter,omitempty"`
	AddListUnsubscribeHeader *bool             `json:"addListUnsubscribeHeader,omitempty"`
}

// SMSRequest is the payload for both the SMS and WhatsApp endpoints.
type SMSRequest struct {
	To   string `json:"to"`
	From string `json:"from"`
	Text string `json:"text"`
}

type CreateDomainRequest struct {
	Name string `json:"name"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

type CreateAudienceRequest struct {
	Name string `json:"name"`
}

type CreateContactRequest struct {
	Email        string `json:"email"`
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	Unsubscribed *bool  `json:"unsubscribed,omitempty"`
}

type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

type CreateTemplateRequest struct {
	Name        string `json:"name"`
	Subject     string `json:"subject,omitempty"`
	HTMLContent string `json:"htmlContent,omitempty"`
}

type UpdateTemplateRequest struct {
	Name        string `json:"name,omitempty"`
	Subject     string `json:"subject,omitempty"`
	HTMLContent string `json:"htmlContent,omitempty"`
}

type CreateSuppressionRequest struct {
	Email  string `json:"email"`
	Reason string `json:"reason,omitempty"`
}

// ─── Responses ────────────────────────────────────────────────────────────

type EmailResponse struct {
	ID string `json:"id"`
}

type SmsResponse struct {
	ID string `json:"id"`
}

type Domain struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type DkimToken struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Type   string `json:"type"`
	Status *bool  `json:"status,omitempty"`
}

type DkimResponse struct {
	Status string      `json:"status"`
	Tokens []DkimToken `json:"tokens"`
}

type VerifyDomainResponse struct {
	Status      string         `json:"status"`
	Message     string         `json:"message,omitempty"`
	Diagnostics map[string]any `json:"diagnostics,omitempty"`
}

type APIKeyCreated struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Token     string `json:"token"`
	CreatedAt string `json:"createdAt"`
}

type APIKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type Audience struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type Contact struct {
	ID           string `json:"id"`
	AudienceID   string `json:"audienceId"`
	Email        string `json:"email"`
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	Unsubscribed *bool  `json:"unsubscribed,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
}

type Webhook struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Secret    string   `json:"secret"`
	Events    []string `json:"events"`
	CreatedAt string   `json:"createdAt,omitempty"`
}

type WebhookDelivery struct {
	ID           string `json:"id"`
	WebhookID    string `json:"webhookId"`
	EventID      string `json:"eventId,omitempty"`
	EventType    string `json:"eventType,omitempty"`
	Attempt      int    `json:"attempt,omitempty"`
	StatusCode   int    `json:"statusCode,omitempty"`
	ResponseBody string `json:"responseBody,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
}

type Template struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Subject     string `json:"subject,omitempty"`
	HTMLContent string `json:"htmlContent,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type Suppression struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Reason    string `json:"reason,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type MessageLog struct {
	ID                string `json:"id"`
	Channel           string `json:"channel"`
	Status            string `json:"status"`
	ToAddress         string `json:"toAddress,omitempty"`
	FromAddress       string `json:"fromAddress,omitempty"`
	Subject           string `json:"subject,omitempty"`
	HTMLBody          string `json:"htmlBody,omitempty"`
	TextBody          string `json:"textBody,omitempty"`
	ProviderMessageID string `json:"providerMessageId,omitempty"`
	CCAddress         string `json:"ccAddress,omitempty"`
	BCCAddress        string `json:"bccAddress,omitempty"`
	ReplyToAddress    string `json:"replyToAddress,omitempty"`
	CreatedAt         string `json:"createdAt,omitempty"`
}
