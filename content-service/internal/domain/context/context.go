package context

import (
	"context"
	"errors"
	"fmt"
	"time"

	Map "github.com/orcaman/concurrent-map"
	"github.com/spf13/cast"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/logger"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/auth"
)

type sessionName string

const (
	AppSession  sessionName = "App_Session"
	UserSession sessionName = "User_Session"
)

type Context struct {
	Map                       Map.ConcurrentMap
	Logger                    logger.Logger
	RequestTime               time.Time
	UserSession               auth.UserSession
	XCorrelationID            string
	AppName, AppVersion, IP   string
	Port                      int
	SrcIP, URL, Method        string
	Header, Request, Response any
	ErrorMessage              string
	ResponseCode              string
	GrpcAuthToken             string
}

func New() *Context {
	return &Context{
		RequestTime: time.Now(),
		Logger:      logger.NewLogger(),
		Map:         Map.New(),
		Header:      map[string]any{},
		Request:     struct{}{},
	}
}

func (s *Context) SetXCorrelationID(xCorrelationID string) *Context {
	s.XCorrelationID = xCorrelationID
	return s
}

func (s *Context) SetMethod(method string) *Context {
	s.Method = method
	return s
}

func (s *Context) SetAppName(appName string) *Context {
	s.AppName = appName
	return s
}

func (s *Context) SetAppVersion(appVersion string) *Context {
	s.AppVersion = appVersion
	return s
}

func (s *Context) SetURL(url string) *Context {
	s.URL = url
	return s
}

func (s *Context) SetIP(ip string) *Context {
	s.IP = ip
	return s
}

func (s *Context) SetPort(port int) *Context {
	s.Port = port
	return s
}

func (s *Context) SetSrcIP(srcIp string) *Context {
	s.SrcIP = srcIp
	return s
}

func (s *Context) SetHeader(header any) *Context {
	s.Header = header
	return s
}

func (s *Context) SetRequest(request any) *Context {
	s.Request = request
	return s
}

func (s *Context) SetRequestTime(request time.Time) *Context {
	s.RequestTime = request
	return s
}

func (s *Context) SetErrorMessage(errorMessage string) *Context {
	s.ErrorMessage = errorMessage
	return s
}

func (s *Context) SetResponseCode(responseCode string) *Context {
	s.ResponseCode = responseCode
	return s
}

func (s *Context) SetGrpcAuthToken(grpcAuthToken string) *Context {
	s.GrpcAuthToken = grpcAuthToken
	return s
}

func (s *Context) SetLogger(l logger.Logger) *Context {
	s.Logger = l
	return s
}

func (s *Context) SetError(err toddlerr.Error) *Context {
	s.ErrorMessage = err.Error()
	s.SetResponseCode(fmt.Sprintf("public: %d, service: %d", err.PublicStatusCode, err.ServiceStatusCode))
	return s
}

func (s *Context) Get(key string) (data any, err error) {
	data, ok := s.Map.Get(key)
	if !ok {
		err = errors.New("not found")
	}
	return
}

func (s *Context) Put(key string, data any) {
	s.Map.Set(key, data)
}

func (s *Context) Lv1(message ...any) {
	s.Logger.Info(s.toContextLogger("Lv1"), "", formatLogs(message...)...)
}

func (s *Context) Lv2(message ...any) time.Time {
	s.Logger.Info(s.toContextLogger("Lv2"), "", formatLogs(message...)...)
	return time.Now()
}

func (s *Context) Lv3(startProcessTime time.Time, message ...any) {
	stop := time.Now()

	msg := formatLogs(message...)
	msg = append(msg, logger.ToField("_process_time", stop.Sub(startProcessTime).Nanoseconds()/1000000))

	s.Logger.Info(s.toContextLogger("Lv3"), "", msg...)
}

func (s *Context) Lv4(message ...any) {
	stop := time.Now()
	rt := stop.Sub(s.RequestTime).Nanoseconds() / 1000000

	msg := formatLogs(message...)
	msg = append(msg, logger.ToField("_response_time", rt))

	s.Logger.Info(s.toContextLogger("Lv4"), "", msg...)
}

func (s *Context) Error(message string, field ...any) {
	s.Logger.Error(s.toContextLogger("Error"), message, formatLogs(field...)...)
}

func formatLogs(message ...any) (logRecord []logger.Field) {
	for index, msg := range message {
		logRecord = append(logRecord, logger.ToField("_message_"+cast.ToString(index), msg))
	}

	return
}

func (s *Context) toContextLogger(tag string) (ctx context.Context) {
	ctxVal := logger.Context{
		ServiceName:    s.AppName,
		ServiceVersion: s.AppVersion,
		ServicePort:    s.Port,
		XCorrelationID: s.XCorrelationID,
		Tag:            tag,
		ReqMethod:      s.Method,
		ReqURI:         s.URL,
		AdditionalData: s.Map.Items(),
		Error:          s.ErrorMessage,
	}
	if tag == "Lv4" {
		ctxVal.Request = logger.ToField("req", s.Request)
		ctxVal.Response = logger.ToField("resp", s.Response)
	}

	ctx = logger.InjectCtx(context.Background(), ctxVal)
	return
}
