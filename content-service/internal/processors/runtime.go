package processors

import (
	"github.com/beka-birhanu/yetbota/content-service/drivers/rabbitmq"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	driverConfig "github.com/beka-birhanu/yetbota/content-service/drivers/config"
	domainComment "github.com/beka-birhanu/yetbota/content-service/internal/domain/comment"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainFeed "github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	domainFollower "github.com/beka-birhanu/yetbota/content-service/internal/domain/follower"
	domainNotification "github.com/beka-birhanu/yetbota/content-service/internal/domain/notification"
	domainPhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/photo"
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
	domainActivity "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeactivity"
	domainPlaceMatch "github.com/beka-birhanu/yetbota/content-service/internal/domain/placematch"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	domainPostSim "github.com/beka-birhanu/yetbota/content-service/internal/domain/postsimilarity"
	domainPostvote "github.com/beka-birhanu/yetbota/content-service/internal/domain/postvote"
	processorsDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	domainUser "github.com/beka-birhanu/yetbota/content-service/internal/domain/user"
)

type Runtime struct {
	Publisher          processorsDomain.Executor
	NewPostHandler     *NewPostHandler
	FeedUpdateHandler  *FeedUpdateHandler
	PostDigestHandler  *PostDigestHandler
	PlaceIngestHandler *PlaceIngestHandler
	AuthorScoringLoop  *AuthorScoringLoop
}

type RuntimeConfig struct {
	RabbitMQClient       *rabbitmq.Client       `validate:"required"`
	PostPhotoRepo        domainPostphoto.Repository `validate:"required"`
	PhotoRepo            domainPhoto.Repository     `validate:"required"`
	Bucket               domainStorage.Bucket       `validate:"required"`
	FollowerRepo         domainFollower.Repository  `validate:"required"`
	PostSimRepo          domainPostSim.Repository   `validate:"required"`
	FeedRepo             domainFeed.Repository      `validate:"required"`
	PostRepo             domainPost.Repository      `validate:"required"`
	CommentRepo          domainComment.Repository   `validate:"required"`
	PostvoteRepo         domainPostvote.Repository  `validate:"required"`
	PlaceRepo            domainPlace.Repository
	ContributionRepo     domainContribution.Repository
	ActivityRepo         domainActivity.Repository
	PlaceMatcher         domainPlaceMatch.Client
	PlaceLinkMinScore    float64
	PlaceLinkRadiusKm    float64
	BatchStore           domainStorage.Set          `validate:"required"`
	SeenCache            domainStorage.Set          `validate:"required"`
	SummarizeDebounce    domainStorage.Set
	SummarizeDebounceSec int64
	SeedBonus            float64                    `validate:"required"`
	QScale               float64                    `validate:"required"`
	Epoch                int64                      `validate:"required"`
	HalfLifeHours        float64                    `validate:"required"`
	MinFeedScore         float64                    `validate:"required"`
	CelebrityThreshold   int64                      `validate:"required"`
	MaxCelebrityFeedSize int64                      `validate:"required"`
	ScoringStream        domainStorage.Stream          `validate:"required"`
	UserRepo             domainUser.Repository         `validate:"required"`
	NotificationSender   domainNotification.Sender     `validate:"required"`
	NotificationRepo     domainNotification.Repository `validate:"required"`
	Badges               []driverConfig.Badge
	AuthorConsumerGroup  string `validate:"required"`
	AuthorBatchSize      int64  `validate:"required"`
	AuthorScoringInput   processorsDomain.AuthorScoringWorkflowInput `validate:"required"`
}

func NewRuntime(cfg *RuntimeConfig) (*Runtime, error) {
	if err := validator.Validate.Struct(cfg); err != nil {
		return nil, err
	}

	newPostAct, err := newNewPostActivity(&newPostActConfig{
		PostPhotoRepo:      cfg.PostPhotoRepo,
		PhotoRepo:          cfg.PhotoRepo,
		Bucket:             cfg.Bucket,
		PostRepo:           cfg.PostRepo,
		NotificationRepo:   cfg.NotificationRepo,
		NotificationSender: cfg.NotificationSender,
		UserRepo:           cfg.UserRepo,
	})
	if err != nil {
		return nil, err
	}

	feedUpdateAct, err := newFeedUpdateAct(&feedUpdateActConfig{
		FollowerRepo:         cfg.FollowerRepo,
		PostSimRepo:          cfg.PostSimRepo,
		FeedRepo:             cfg.FeedRepo,
		PostRepo:             cfg.PostRepo,
		PostvoteRepo:         cfg.PostvoteRepo,
		BatchStore:           cfg.BatchStore,
		SeenCache:            cfg.SeenCache,
		SeedBonus:            cfg.SeedBonus,
		QScale:               cfg.QScale,
		Epoch:                cfg.Epoch,
		HalfLifeHours:        cfg.HalfLifeHours,
		MinFeedScore:         cfg.MinFeedScore,
		CelebrityThreshold:   cfg.CelebrityThreshold,
		MaxCelebrityFeedSize: cfg.MaxCelebrityFeedSize,
	})
	if err != nil {
		return nil, err
	}

	authorScoringAct, err := newAuthorScoringAct(&authorScoringActConfig{
		Stream:             cfg.ScoringStream,
		PostRepo:           cfg.PostRepo,
		UserRepo:           cfg.UserRepo,
		NotificationSender: cfg.NotificationSender,
		Badges:             cfg.Badges,
		ConsumerGroup:      cfg.AuthorConsumerGroup,
		BatchSize:          cfg.AuthorBatchSize,
	})
	if err != nil {
		return nil, err
	}

	publisher := rabbitmq.NewPublisher(cfg.RabbitMQClient)
	consumer := rabbitmq.NewConsumer(cfg.RabbitMQClient)

	jobPublisher, err := NewJobPublisher(&JobPublisherConfig{
		Publisher:            publisher,
		PostRepo:             cfg.PostRepo,
		CommentRepo:          cfg.CommentRepo,
		SummarizeDebounce:    cfg.SummarizeDebounce,
		SummarizeDebounceSec: cfg.SummarizeDebounceSec,
	})
	if err != nil {
		return nil, err
	}

	newPostHandler, err := NewNewPostHandler(&NewPostHandlerConfig{
		NewPostActivity: newPostAct,
		PostRepo:        cfg.PostRepo,
		Consumer:        consumer,
		Publisher:       publisher,
	})
	if err != nil {
		return nil, err
	}

	feedUpdateHandler, err := NewFeedUpdateHandler(&FeedUpdateHandlerConfig{
		FeedUpdateActivity: feedUpdateAct,
		Consumer:           consumer,
	})
	if err != nil {
		return nil, err
	}

	authorScoringLoop, err := NewAuthorScoringLoop(&AuthorScoringLoopConfig{
		Activity: authorScoringAct,
		Input:    cfg.AuthorScoringInput,
	})
	if err != nil {
		return nil, err
	}

	var postDigestHandler *PostDigestHandler
	var placeIngestHandler *PlaceIngestHandler
	if cfg.PlaceRepo != nil && cfg.ContributionRepo != nil && cfg.ActivityRepo != nil {
		postDigestHandler, err = NewPostDigestHandler(&PostDigestHandlerConfig{
			PostRepo:         cfg.PostRepo,
			PlaceRepo:        cfg.PlaceRepo,
			ContributionRepo: cfg.ContributionRepo,
			ActivityRepo:     cfg.ActivityRepo,
			Consumer:         rabbitmq.NewConsumer(cfg.RabbitMQClient),
			Publisher:        publisher,
			Executor:         jobPublisher,
			PlaceMatcher:     cfg.PlaceMatcher,
			LinkRadiusKm:     cfg.PlaceLinkRadiusKm,
			MinMatchScore:    cfg.PlaceLinkMinScore,
		})
		if err != nil {
			return nil, err
		}
		placeIngestHandler, err = NewPlaceIngestHandler(&PlaceIngestHandlerConfig{
			PlaceRepo:        cfg.PlaceRepo,
			ContributionRepo: cfg.ContributionRepo,
			Publisher:        publisher,
			Executor:         jobPublisher,
			Consumer:         rabbitmq.NewConsumer(cfg.RabbitMQClient),
		})
		if err != nil {
			return nil, err
		}
	}

	return &Runtime{
		Publisher:          jobPublisher,
		NewPostHandler:     newPostHandler,
		FeedUpdateHandler:  feedUpdateHandler,
		PostDigestHandler:  postDigestHandler,
		PlaceIngestHandler: placeIngestHandler,
		AuthorScoringLoop:  authorScoringLoop,
	}, nil
}
