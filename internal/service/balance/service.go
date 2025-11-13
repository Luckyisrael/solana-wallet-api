package balance

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "time"

    solanago "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
    "github.com/Luckyisrael/solana-wallet-api/internal/api/dto"
    "github.com/Luckyisrael/solana-wallet-api/internal/redis"
    "github.com/Luckyisrael/solana-wallet-api/internal/solana"

    "go.uber.org/zap"
)

const (
	cacheTTL      = 5 * time.Minute
	rateLimitKey  = "rate:balance:%s"
	rateLimitTTL  = time.Second
	rateLimitMax  = 10
)

type Service struct {
    solanaClient *solana.Client
    redisClient  *redis.Client
    logger       *zap.Logger
}

func NewService(solanaClient *solana.Client, redisClient *redis.Client) *Service {
	logger, _ := zap.NewProduction()
	return &Service{
		solanaClient: solanaClient,
		redisClient:  redisClient,
		logger:       logger,
	}
}

func (s *Service) GetBalance(ctx context.Context, addressStr string) (*dto.BalanceResponse, error) {
    address, err := solanago.PublicKeyFromBase58(addressStr)
    if err != nil {
        return nil, err
    }

	// Rate limiting
	if err := s.rateLimit(ctx, addressStr); err != nil {
		return nil, err
	}

	// Cache key
	cacheKey := "balance:" + addressStr

	// Try cache
	var cached dto.BalanceResponse
	if err := s.redisClient.GetJSON(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	// Fetch from RPC
	resp, err := s.fetchFromRPC(ctx, address)
	if err != nil {
		return nil, err
	}

	// Cache result
	if cacheErr := s.redisClient.SetJSON(ctx, cacheKey, resp, cacheTTL); cacheErr != nil {
		s.logger.Warn("Failed to cache balance", zap.Error(cacheErr))
	}

	return resp, nil
}

func (s *Service) rateLimit(ctx context.Context, address string) error {
	key := fmt.Sprintf(rateLimitKey, address)
	count, err := s.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		s.redisClient.Expire(ctx, key, rateLimitTTL)
	}
	if count > rateLimitMax {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

func (s *Service) fetchFromRPC(ctx context.Context, address solanago.PublicKey) (*dto.BalanceResponse, error) {
    // Get SOL balance
    solBalance, err := s.solanaClient.GetBalance(ctx, address, rpc.CommitmentFinalized)
    if err != nil {
        return nil, err
    }
    solAmount := float64(solBalance.Value) / 1e9

    // Get token accounts
    tokenAccounts, err := s.solanaClient.GetTokenAccountsByOwner(
        ctx,
        address,
        &rpc.GetTokenAccountsConfig{},
        &rpc.GetTokenAccountsOpts{Encoding: solanago.EncodingBase64},
    )
    // Best-effort: if tokens cannot be fetched, continue with SOL only
    if err != nil {
        return &dto.BalanceResponse{Address: address.String(), SOL: fmt.Sprintf("%.9f", solAmount), Tokens: []dto.TokenBalance{}}, nil
    }

	tokens := make([]dto.TokenBalance, 0)
	for _, acc := range tokenAccounts.Value {
		var data struct {
			Parsed struct {
				Info struct {
					Mint        string `json:"mint"`
					TokenAmount struct {
						Amount         string `json:"amount"`
						Decimals       uint8  `json:"decimals"`
						UIAmount       *float64 `json:"uiAmount"`
						UIAmountString string `json:"uiAmountString"`
					} `json:"tokenAmount"`
				} `json:"info"`
			} `json:"parsed"`
		}

		if err := json.Unmarshal(acc.Account.Data.GetBinary(), &data); err != nil {
			continue
		}

		mint := data.Parsed.Info.Mint
		amount := data.Parsed.Info.TokenAmount.Amount
		decimals := data.Parsed.Info.TokenAmount.Decimals
		uiAmount := 0.0
		if data.Parsed.Info.TokenAmount.UIAmount != nil {
			uiAmount = *data.Parsed.Info.TokenAmount.UIAmount
		}

		// Simple token metadata (SOL, USDC, etc.)
		symbol, name := getTokenMetadata(mint)

		humanAmount := ""
		if decimals > 0 {
			humanAmount = formatAmount(amount, decimals)
		}

		tokens = append(tokens, dto.TokenBalance{
			Mint:     mint,
			Symbol:   symbol,
			Name:     name,
			Amount:   humanAmount,
			Decimals: decimals,
			UIAmount: uiAmount,
		})
	}

    return &dto.BalanceResponse{
        Address: address.String(),
        SOL:     fmt.Sprintf("%.9f", solAmount),
        Tokens:  tokens,
    }, nil
}

// getTokenMetadata - hardcoded for common tokens (expand later)
func getTokenMetadata(mint string) (symbol, name string) {
	known := map[string][2]string{
		"So11111111111111111111111111111111111111112": {"SOL", "Wrapped SOL"},
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyK9u7u": {"USDC", "USD Coin"},
	}
	if v, ok := known[mint]; ok {
		return v[0], v[1]
	}
	return "UNKNOWN", "Unknown Token"
}

// formatAmount converts raw amount to human-readable string
func formatAmount(amount string, decimals uint8) string {
	if decimals == 0 {
		return amount
	}
	if len(amount) <= int(decimals) {
		return "0." + strings.Repeat("0", int(decimals)-len(amount)) + amount
	}
	whole := amount[:len(amount)-int(decimals)]
	fraction := amount[len(amount)-int(decimals):]
	fraction = strings.TrimRight(fraction, "0")
	if fraction == "" {
		return whole
	}
	return whole + "." + fraction
}
