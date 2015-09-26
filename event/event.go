package event

type Message struct {
	Body          string `json:"body"`
	Msgtype       string `json:"msgtype"`
	FormattedBody string `json:"formatted_body,omitempty"`
	Format        string `json:"format,omitempty"`
	URL           string `json:"url,omitempty"`
}

type Member struct {
	Membership string `json:"membership"`
}
