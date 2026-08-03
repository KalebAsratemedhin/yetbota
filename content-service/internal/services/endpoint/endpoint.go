package endpoint

import (
	"github.com/go-kit/kit/endpoint"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	adminSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/admin"
	commentSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/comment"
	feedSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/feed"
	moderationSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/moderation"
	notificationSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/notification"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
	contributionSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/contribution"
	placeSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/place"
	placefeedSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/placefeed"
)

type Endpoints struct {
	PostAdd    endpoint.Endpoint
	PostRead   endpoint.Endpoint
	PostUpdate endpoint.Endpoint
	PostVote   endpoint.Endpoint
	PostList   endpoint.Endpoint
	PostSave   endpoint.Endpoint
	PostUnsave endpoint.Endpoint

	PlaceAdd           endpoint.Endpoint
	PlaceRead          endpoint.Endpoint
	PlaceUpdate        endpoint.Endpoint
	PlaceList          endpoint.Endpoint
	PlaceBookmark      endpoint.Endpoint
	PlaceUnbookmark    endpoint.Endpoint
	PlaceConfirm       endpoint.Endpoint
	PlaceReportProblem endpoint.Endpoint
	PlaceUpdateSummary endpoint.Endpoint

	ContributionAdd               endpoint.Endpoint
	ContributionAddAnswer         endpoint.Endpoint
	ContributionVote              endpoint.Endpoint
	ContributionListByPlace       endpoint.Endpoint
	ContributionAcceptAnswer      endpoint.Endpoint
	ContributionListOpenQuestions endpoint.Endpoint

	PlaceFeedList endpoint.Endpoint

	CommentAdd    endpoint.Endpoint
	CommentRead   endpoint.Endpoint
	CommentList   endpoint.Endpoint
	CommentDelete endpoint.Endpoint
	CommentVote   endpoint.Endpoint

	FeedGet        endpoint.Endpoint
	FeedMarkViewed endpoint.Endpoint

	NotificationList       endpoint.Endpoint
	NotificationMarkAsRead endpoint.Endpoint
	NotificationDelete     endpoint.Endpoint

	ReportCreate        endpoint.Endpoint
	ModerationListCases endpoint.Endpoint
	ModerationGetCase   endpoint.Endpoint
	ModerationAct       endpoint.Endpoint

	AdminOverviewStats  endpoint.Endpoint
	AdminOverviewGrowth endpoint.Endpoint
	AdminUserStats      endpoint.Endpoint
	AdminSystemAudit    endpoint.Endpoint
}

type Config struct {
	PostService         postSvc.Service         `validate:"required"`
	PlaceService        placeSvc.Service        `validate:"required"`
	ContributionService contributionSvc.Service `validate:"required"`
	PlaceFeedService    placefeedSvc.Service    `validate:"required"`
	CommentService      commentSvc.Service      `validate:"required"`
	FeedService         feedSvc.Service         `validate:"required"`
	NotificationService notificationSvc.Service `validate:"required"`
	ModerationService   moderationSvc.Service   `validate:"required"`
	AdminService        adminSvc.Service        `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewEndpoints(c *Config) (*Endpoints, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &Endpoints{
		PostAdd:    makePostAddEndpoint(c.PostService),
		PostRead:   makePostReadEndpoint(c.PostService),
		PostUpdate: makePostUpdateEndpoint(c.PostService),
		PostVote:   makePostVoteEndpoint(c.PostService),
		PostList:   makePostListEndpoint(c.PostService),
		PostSave:   makePostSaveEndpoint(c.PostService),
		PostUnsave: makePostUnsaveEndpoint(c.PostService),

		PlaceAdd:           makePlaceAddEndpoint(c.PlaceService),
		PlaceRead:          makePlaceReadEndpoint(c.PlaceService),
		PlaceUpdate:        makePlaceUpdateEndpoint(c.PlaceService),
		PlaceList:          makePlaceListEndpoint(c.PlaceService),
		PlaceBookmark:      makePlaceBookmarkEndpoint(c.PlaceService),
		PlaceUnbookmark:    makePlaceUnbookmarkEndpoint(c.PlaceService),
		PlaceConfirm:       makePlaceConfirmEndpoint(c.PlaceService),
		PlaceReportProblem: makePlaceReportProblemEndpoint(c.PlaceService),
		PlaceUpdateSummary: makePlaceUpdateSummaryEndpoint(c.PlaceService),

		ContributionAdd:               makeContributionAddEndpoint(c.ContributionService),
		ContributionAddAnswer:         makeContributionAddAnswerEndpoint(c.ContributionService),
		ContributionVote:              makeContributionVoteEndpoint(c.ContributionService),
		ContributionListByPlace:       makeContributionListByPlaceEndpoint(c.ContributionService),
		ContributionAcceptAnswer:      makeContributionAcceptAnswerEndpoint(c.ContributionService),
		ContributionListOpenQuestions: makeContributionListOpenQuestionsEndpoint(c.ContributionService),

		PlaceFeedList: makePlaceFeedListEndpoint(c.PlaceFeedService),

		CommentAdd:    makeCommentAddEndpoint(c.CommentService),
		CommentRead:   makeCommentReadEndpoint(c.CommentService),
		CommentList:   makeCommentListEndpoint(c.CommentService),
		CommentDelete: makeCommentDeleteEndpoint(c.CommentService),
		CommentVote:   makeCommentVoteEndpoint(c.CommentService),

		FeedGet:        makeFeedGetEndpoint(c.FeedService),
		FeedMarkViewed: makeFeedMarkViewedEndpoint(c.FeedService),

		NotificationList:       makeNotificationListEndpoint(c.NotificationService),
		NotificationMarkAsRead: makeNotificationMarkAsReadEndpoint(c.NotificationService),
		NotificationDelete:     makeNotificationDeleteEndpoint(c.NotificationService),

		ReportCreate:        makeReportCreateEndpoint(c.ModerationService),
		ModerationListCases: makeModerationListCasesEndpoint(c.ModerationService),
		ModerationGetCase:   makeModerationGetCaseEndpoint(c.ModerationService),
		ModerationAct:       makeModerationActEndpoint(c.ModerationService),

		AdminOverviewStats:  makeAdminOverviewStatsEndpoint(c.AdminService),
		AdminOverviewGrowth: makeAdminOverviewGrowthEndpoint(c.AdminService),
		AdminUserStats:      makeAdminUserStatsEndpoint(c.AdminService),
		AdminSystemAudit:    makeAdminSystemAuditEndpoint(c.AdminService),
	}, nil
}
