package aliyun

import (
	"context"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"

	"aliyun-ip-manager/internal/config"
	"aliyun-ip-manager/internal/logger"
)

type Client struct {
	accessKeyID     string
	accessKeySecret string
	region          string
	sdkClient       *sdk.Client
}

type ClientOption func(*Client)

func WithRegion(region string) ClientOption {
	return func(c *Client) {
		c.region = region
	}
}

func WithCredentials(accessKeyID, accessKeySecret string) ClientOption {
	return func(c *Client) {
		c.accessKeyID = accessKeyID
		c.accessKeySecret = accessKeySecret
	}
}

func NewClient(cfg *config.Config, opts ...ClientOption) (*Client, error) {
	client := &Client{
		accessKeyID:     cfg.AccessKeyID,
		accessKeySecret: cfg.AccessKeySecret,
		region:          cfg.Region,
	}

	for _, opt := range opts {
		opt(client)
	}

	if client.accessKeyID == "" || client.accessKeySecret == "" {
		return nil, fmt.Errorf("access key id and secret are required")
	}

	if client.region == "" {
		return nil, fmt.Errorf("region is required")
	}

	cred := credentials.NewAccessKeyCredential(client.accessKeyID, client.accessKeySecret)
	sdkClient, err := sdk.NewClientWithOptions(client.region, client.getConfig(), cred)
	if err != nil {
		return nil, fmt.Errorf("failed to create aliyun client: %w", err)
	}

	client.sdkClient = sdkClient
	return client, nil
}

func (c *Client) getConfig() *sdk.Config {
	return sdk.NewConfig().
		WithTimeout(time.Second * 30).
		WithAutoRetry(true).
		WithMaxRetryTime(3)
}

func (c *Client) DoRequest(ctx context.Context, request requests.AcsRequest, response responses.AcsResponse) error {
	if request.GetScheme() == "" {
		request.SetScheme("HTTPS")
	}

	logger.Info("making aliyun api request",
		"region", c.region)

	err := c.doRequestWithRetry(ctx, request, response, 3)

	var action, product string
	func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Warn("recovered from panic when getting action/product", "panic", r)
			}
		}()
		action = request.GetActionName()
		product = request.GetProduct()
	}()

	if err != nil {
		logger.Error("aliyun api request failed",
			"region", c.region,
			"action", action,
			"product", product,
			"error", err)
		return err
	}

	var httpStatus int
	func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Warn("recovered from panic when getting http status", "panic", r)
				httpStatus = 0
			}
		}()
		httpStatus = response.GetHttpStatus()
	}()

	logger.Info("aliyun api request succeeded",
		"region", c.region,
		"action", action,
		"product", product,
		"response_status", httpStatus)

	if commonResp, ok := response.(*responses.CommonResponse); ok {
		var content []byte
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Warn("recovered from panic when getting response content", "panic", r)
					content = nil
				}
			}()
			content = commonResp.GetHttpContentBytes()
		}()
		if content != nil {
			logger.Debug("api response content",
				"action", action,
				"content", string(content))
		}
	}

	return nil
}

func (c *Client) doRequestWithRetry(ctx context.Context, request requests.AcsRequest, response responses.AcsResponse, maxRetries int) error {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Warn("recovered from panic in sdkClient.DoAction", "panic", r)
					err = fmt.Errorf("panic in sdkClient.DoAction: %v", r)
				}
			}()
			err = c.sdkClient.DoAction(request, response)
		}()

		if err == nil {
			return nil
		}

		lastErr = err

		if !isRetryableError(err) {
			return err
		}

		logger.Warn("retrying aliyun api request",
			"attempt", attempt+1,
			"max_retries", maxRetries,
			"error", err)

		time.Sleep(backoff(attempt))
	}

	return fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

func (c *Client) GetRegion() string {
	return c.region
}

func (c *Client) GetSDKClient() *sdk.Client {
	return c.sdkClient
}

func (c *Client) GetAccessKeyID() string {
	return c.accessKeyID
}

func (c *Client) GetAccessKeySecret() string {
	return c.accessKeySecret
}
