package solana

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type Client struct {
    *rpc.Client
}

func NewClient(endpoint string) *Client {
	return &Client{rpc.New(endpoint)}
}

// GetLatestBlockhash returns the latest finalized blockhash
func (c *Client) GetLatestBlockhash(ctx context.Context) (solana.Hash, error) {
    resp, err := c.Client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
    if err != nil {
        return solana.Hash{}, err
    }
    return resp.Value.Blockhash, nil
}
