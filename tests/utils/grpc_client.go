package utils

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClient represents a client for interacting with the gRPC API
type GRPCClient struct {
	conn *grpc.ClientConn
}

// NewGRPCClient creates a new gRPC client for the specified address
func NewGRPCClient(address string) (*GRPCClient, error) {
	// Set up a connection to the gRPC server
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &GRPCClient{
		conn: conn,
	}, nil
}

// Close closes the gRPC connection
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetConnection returns the underlying gRPC connection
func (c *GRPCClient) GetConnection() *grpc.ClientConn {
	return c.conn
}

// ExecuteWithRetry executes the provided function with retry logic
func (c *GRPCClient) ExecuteWithRetry(ctx context.Context, fn func() error, maxRetries int, retryDelay time.Duration) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := fn(); err != nil {
				lastErr = err
				time.Sleep(retryDelay)
				continue
			}
			return nil
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}
