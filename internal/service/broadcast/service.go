package broadcast

import (
    "context"
    "crypto/sha256"
    "encoding/base64"
    "encoding/hex"
    "fmt"
    "time"

    bin "github.com/gagliardetto/binary"
    solanago "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
    "github.com/Luckyisrael/solana-wallet-api/internal/api/dto"
    "github.com/Luckyisrael/solana-wallet-api/internal/redis"
    "github.com/Luckyisrael/solana-wallet-api/internal/solana"
    "go.uber.org/zap"
)

const (
	idempotencyTTL = 24 * time.Hour
	broadcastTTL   = 30 * time.Second
	pollInterval   = 2 * time.Second
	rateLimitKey   = "rate:broadcast"
	rateLimitMax   = 5
	rateLimitWin   = time.Second
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

func (s *Service) Broadcast(ctx context.Context, req *dto.BroadcastRequest) (*dto.BroadcastResponse, error) {
	// Rate limiting
	if err := s.rateLimit(ctx); err != nil {
		return nil, err
	}

    // Decode tx
    txBytes, err := base64.StdEncoding.DecodeString(req.SignedTxBase64)
    if err != nil {
        return nil, fmt.Errorf("invalid base64")
    }

    key := req.IdempotencyKey
    if key == "" {
        sum := sha256.Sum256(txBytes)
        key = hex.EncodeToString(sum[:])
    }

    var cached dto.BroadcastResponse
    if s.redisClient.GetJSON(ctx, "idempotency:"+key, &cached) == nil {
        return &cached, nil
    }

    tx, err := solanago.TransactionFromDecoder(bin.NewBinDecoder(txBytes))
	if err != nil {
		return nil, fmt.Errorf("invalid transaction")
	}

	// Send
    sig, err := s.solanaClient.SendTransaction(ctx, tx)
	if err != nil {
		s.logger.Error("Broadcast failed", zap.Error(err), zap.String("sig", sig.String()))
		return nil, fmt.Errorf("broadcast failed: %v", err)
	}

	// Poll for confirmation
	resp := &dto.BroadcastResponse{Signature: sig.String()}
	status, slot, err := s.pollConfirmation(ctx, sig)
	if err != nil {
		resp.Status = "failed"
		resp.Error = err.Error()
	} else {
		resp.Status = status
		resp.Slot = slot
	}

    s.redisClient.SetJSON(ctx, "idempotency:"+key, resp, idempotencyTTL)

	s.logger.Info("Transaction broadcast",
		zap.String("signature", sig.String()),
		zap.String("status", resp.Status),
	)

	return resp, nil
}

func (s *Service) rateLimit(ctx context.Context) error {
	count, err := s.redisClient.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		s.redisClient.Expire(ctx, rateLimitKey, rateLimitWin)
	}
	if count > rateLimitMax {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

func (s *Service) pollConfirmation(ctx context.Context, sig solanago.Signature) (string, uint64, error) {
	deadline := time.Now().Add(broadcastTTL)
	for time.Now().Before(deadline) {
        status, err := s.solanaClient.GetSignatureStatuses(ctx, true, sig)
		if err != nil {
			time.Sleep(pollInterval)
			continue
		}
		if len(status.Value) == 0 || status.Value[0] == nil {
			time.Sleep(pollInterval)
			continue
		}
		st := status.Value[0]
        if st.ConfirmationStatus == rpc.ConfirmationStatusFinalized {
            return "confirmed", st.Slot, nil
        }
		if st.Err != nil {
			return "", 0, fmt.Errorf("tx failed: %v", st.Err)
		}
		time.Sleep(pollInterval)
	}
	return "pending", 0, nil
}
