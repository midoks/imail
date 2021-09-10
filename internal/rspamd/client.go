package rspamd

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"strings"
)

const (
	checkV2Endpoint   = "checkv2"
	fuzzyAddEndpoint  = "fuzzyadd"
	fuzzyDelEndpoint  = "fuzzydel"
	learnSpamEndpoint = "learnspam"
	learnHamEndpoint  = "learnham"
	pingEndpoint      = "ping"
)

// Client is a rspamd HTTP client.
type Client interface {
	Check(context.Context, *Email) (*CheckResponse, error)
	LearnSpam(context.Context, *Email) (*LearnResponse, error)
	LearnHam(context.Context, *Email) (*LearnResponse, error)
	FuzzyAdd(context.Context, *Email) (*LearnResponse, error)
	FuzzyDel(context.Context, *Email) (*LearnResponse, error)
	Ping(context.Context) (PingResponse, error)
}

type client struct {
	client   *resty.Client
	password string
}

var _ Client = &client{}

// CheckResponse encapsulates the response of Check.
type CheckResponse struct {
	Score     float64               `json:"score"`
	MessageID string                `json:"message-id"`
	Symbols   map[string]SymbolData `json:"symbols"`
}

// LearnResponse encapsulates the response of LearnSpam, LearnHam, FuzzyAdd, FuzzyDel.
type LearnResponse struct {
	Success bool `json:"success"`
}

// PingResponse encapsulates the response of Ping.
type PingResponse string

// Option is a function that configures the rspamd client.
type Option func(*client) error

type UnexpectedResponseError struct {
	Status int
}

// New returns a client.
// It takes the url of a rspamd instance, and configures the client with Options which are closures.
func New(url string, options ...Option) *client {
	client := &client{
		client: resty.New().SetHostURL(url),
	}

	for _, option := range options {
		err := option(client)
		if err != nil {
			log.Fatal("failed to configure client")
		}
	}

	return client
}

func (c *client) SetAuth(password string) {
	c.password = password
}

// Check scans an email, returning a spam score and list of symbols.
func (c *client) Check(ctx context.Context, e *Email) (*CheckResponse, error) {
	result := &CheckResponse{}
	req := c.makeEmailRequest(ctx, e).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, checkV2Endpoint)
	return result, err
}

// LearnSpam trains rspamd's Bayesian classifier by marking an email as spam.
func (c *client) LearnSpam(ctx context.Context, e *Email) (*LearnResponse, error) {
	result := &LearnResponse{}
	req := c.makeEmailRequest(ctx, e).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, learnSpamEndpoint)
	return result, err
}

// LearnSpam trains rspamd's Bayesian classifier by marking an email as ham.
func (c *client) LearnHam(ctx context.Context, e *Email) (*LearnResponse, error) {
	result := &LearnResponse{}
	req := c.makeEmailRequest(ctx, e).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, learnHamEndpoint)
	return result, err
}

// FuzzyAdd adds an email to fuzzy storage.
func (c *client) FuzzyAdd(ctx context.Context, e *Email) (*LearnResponse, error) {
	result := &LearnResponse{}
	req := c.makeEmailRequest(ctx, e).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, fuzzyAddEndpoint)
	return result, err
}

// FuzzyAdd removes an email from fuzzy storage.
func (c *client) FuzzyDel(ctx context.Context, e *Email) (*LearnResponse, error) {
	result := &LearnResponse{}
	req := c.makeEmailRequest(ctx, e).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, fuzzyDelEndpoint)
	return result, err
}

// Ping pings the client's rspamd instance.
func (c *client) Ping(ctx context.Context) (PingResponse, error) {
	var result PingResponse
	_, err := c.sendRequest(c.client.R().SetContext(ctx).SetResult(result), resty.MethodGet, pingEndpoint)
	return result, err
}

func (c *client) makeEmailRequest(ctx context.Context, e *Email) *resty.Request {
	headers := map[string]string{}
	if e.queueID != "" {
		headers["Queue-ID"] = e.queueID
	}
	if e.options.flag != 0 {
		headers["Flag"] = fmt.Sprintf("%d", e.options.flag)
	}
	if e.options.weight != 0.0 {
		headers["Weight"] = fmt.Sprintf("%f", e.options.weight)
	}
	return c.client.R().
		SetContext(ctx).
		SetHeaders(headers).
		SetBody(e.message)
}

func (c *client) sendRequest(req *resty.Request, method, url string) (*resty.Response, error) {
	if !strings.EqualFold(c.password, "") {
		url = fmt.Sprintf("%s?password=%s", url, c.password)
	}

	res, err := req.Execute(method, url)

	if err != nil {
		return nil, fmt.Errorf("executing request: %q", err)
	}
	if res.StatusCode() != http.StatusOK {
		return nil, &UnexpectedResponseError{Status: res.StatusCode()}
	}

	return res, nil
}

// Credentials sets the credentials passed in parameters.
// It returns an Option which is used to configure the client.
func Credentials(username string, password string) Option {
	return func(c *client) error {
		c.client.SetBasicAuth(username, password).SetHeader("User", username)
		return nil
	}
}

func (e *UnexpectedResponseError) Error() string {
	return fmt.Sprintf("Unexpected response code: %d", e.Status)
}

// IsNotFound returns true if a request returned a 404. This helps discern a known issue with rspamd's /checkv2 endpoint.
func IsNotFound(err error) bool {
	var errResp *UnexpectedResponseError
	return errors.As(err, &errResp) && errResp.Status == http.StatusNotFound
}

// IsAlreadyLearnedError returns true if a request returns 208, which can happen if rspamd detects a message has already been learned as SPAM/HAM.
// This can allow clients to gracefully handle this use case.
func IsAlreadyLearnedError(err error) bool {
	var errResp *UnexpectedResponseError
	return errors.As(err, &errResp) && errResp.Status == http.StatusAlreadyReported
}
