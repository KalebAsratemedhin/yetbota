package messaging

const (
	ExchangeName       = "yetbota"
	DeadLetterExchange = "yetbota.dlx"

	QueueNewPost        = "content.new_post"
	QueuePlaceIngest    = "content.place_ingest"
	QueueFeedUpdate     = "content.feed_update"
	QueuePostDigest     = "content.post_digest"
	QueuePlaceSummarize = "place.summarize"
	QueueAIIngest       = "ai.ingest"

	QueueNewPostDLQ        = "content.new_post.dlq"
	QueuePlaceIngestDLQ    = "content.place_ingest.dlq"
	QueueFeedUpdateDLQ     = "content.feed_update.dlq"
	QueuePostDigestDLQ     = "content.post_digest.dlq"
	QueuePlaceSummarizeDLQ = "place.summarize.dlq"
	QueueAIIngestDLQ       = "ai.ingest.dlq"
)

var WorkQueues = []string{
	QueueNewPost,
	QueuePlaceIngest,
	QueueFeedUpdate,
	QueuePostDigest,
	QueuePlaceSummarize,
	QueueAIIngest,
}
