package messaging

type NewPostMessage struct {
	PostID string `json:"post_id"`
}

type PostDigestMessage struct {
	PostID string `json:"post_id"`
}

type PlaceIngestMessage struct {
	PlaceID string `json:"place_id"`
}

type ContributionIngestMessage struct {
	ContributionID string `json:"contribution_id"`
}

type PlaceSummarizeMessage struct {
	PlaceID string `json:"place_id"`
}

type FeedUpdateMessage struct {
	PostID       string `json:"post_id"`
	InteractorID string `json:"interactor_id,omitempty"`
}

type IngestMessage struct {
	ContentID string   `json:"content_id"`
	PlaceID   string   `json:"place_id,omitempty"`
	Kind      string   `json:"kind"`
	UserID    string   `json:"user_id"`
	Text      string   `json:"text"`
	ParentID  string   `json:"parent_id,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Category  string   `json:"category,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Context   string   `json:"context,omitempty"`
}

type IngestResultMessage struct {
	ContentID   string `json:"content_id"`
	Kind        string `json:"kind"`
	Status      string `json:"status"`
	DuplicateOf string `json:"duplicate_of,omitempty"`
	ErrorCode   string `json:"error_code,omitempty"`
	ProcessedAt string `json:"processed_at"`
}
