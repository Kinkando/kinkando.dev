package line

type WebhookBody struct {
	Destination string  `json:"destination"`
	Events      []Event `json:"events"`
}

type Event struct {
	Type            string          `json:"type"`
	Mode            string          `json:"mode"`
	Timestamp       int64           `json:"timestamp"`
	WebhookEventID  string          `json:"webhookEventId"`
	DeliveryContext DeliveryContext `json:"deliveryContext"`
	Source          Source          `json:"source"`
	ReplyToken      string          `json:"replyToken,omitempty"`
	Message         *Message        `json:"message,omitempty"`
}

type DeliveryContext struct {
	IsRedelivery bool `json:"isRedelivery"`
}

type Source struct {
	Type    string `json:"type"`
	UserID  string `json:"userId,omitempty"`
	GroupID string `json:"groupId,omitempty"`
	RoomID  string `json:"roomId,omitempty"`
}

type Message struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	QuoteToken      string `json:"quoteToken,omitempty"`
	Text            string `json:"text,omitempty"`
	FileName        string `json:"fileName,omitempty"`
	FileSize        int64  `json:"fileSize,omitempty"`
	Title           string `json:"title,omitempty"`
	Address         string `json:"address,omitempty"`
	Duration        int64  `json:"duration,omitempty"`
	ContentProvider string `json:"contentProvider,omitempty"`
}
