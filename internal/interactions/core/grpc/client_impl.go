package grpc

import (
	"context"

	api "github.com/Alp4ka/classifier-api"
	"github.com/Alp4ka/classifier-telegram/internal/interactions/core"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Config struct {
	GRPCAddr string
}

type Client struct {
	grpcClient api.GWManagerServiceClient
}

var _ core.Client = (*Client)(nil)

func NewClient(cfg Config) (*Client, error) {
	conn, err := grpc.Dial(cfg.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	grpcClient := api.NewGWManagerServiceClient(conn)

	return &Client{grpcClient: grpcClient}, nil
}

func (c *Client) AcquireSession(ctx context.Context, agent, gateway string) (uuid.UUID, error) {
	acquireResp, err := c.grpcClient.AcquireSession(ctx, &api.AcquireSessionRequest{
		Agent:   agent,
		Gateway: gateway,
	})
	if err != nil {
		return uuid.Nil, err
	}

	ret, err := uuid.Parse(acquireResp.SessionId)
	if err != nil {
		return uuid.Nil, err
	}

	return ret, nil
}

func (c *Client) ReleaseSession(ctx context.Context, sessionID uuid.UUID) error {
	releaseReq := &api.ReleaseSessionRequest{
		SessionId: sessionID.String(),
	}
	_, err := c.grpcClient.ReleaseSession(ctx, releaseReq)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Process(ctx context.Context, sessionID uuid.UUID) (core.Handler, error) {
	const sessionIDHeader = "session-id"

	md := metadata.Pairs(sessionIDHeader, sessionID.String())
	ctx = metadata.NewOutgoingContext(ctx, md)
	stream, err := c.grpcClient.Process(ctx)
	if err != nil {
		return nil, err
	}

	inputChan := make(chan *api.ProcessRequest, 1)
	outputChan := make(chan *api.ProcessResponse, 1)

	h := newHandler(inputChan, outputChan)

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer func() {
			cancel()
			close(inputChan)
		}()

		for {
			select {
			case <-h.Done():
				return
			case req := <-inputChan:
				err := stream.Send(req)
				if err != nil {
					h.CloseWithError(err)
				}
			}
		}
	}()

	go func() {
		defer func() {
			cancel()
			close(outputChan)
		}()

		for {
			response, err := stream.Recv()
			if err != nil {
				h.CloseWithError(err)
				return
			}

			outputChan <- response
		}
	}()

	return h, nil
}
