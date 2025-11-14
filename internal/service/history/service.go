package history

import (
    "context"

    solanago "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
    "github.com/Luckyisrael/solana-wallet-api/internal/solana"
)

type Service struct {
    solanaClient *solana.Client
}

func NewService(solanaClient *solana.Client) *Service {
    return &Service{solanaClient: solanaClient}
}

type Item struct {
    Signature string
    Slot      uint64
    Err       any
    Status    rpc.ConfirmationStatusType
}

type Response struct {
    Items []Item
}

func (s *Service) GetHistory(ctx context.Context, address solanago.PublicKey, limit int, before string) (*Response, error) {
    sigs, err := s.solanaClient.GetSignaturesForAddress(ctx, address)
    if err != nil {
        return nil, err
    }
    items := make([]Item, 0, len(sigs))
    for _, sgn := range sigs {
        items = append(items, Item{Signature: sgn.Signature.String(), Slot: sgn.Slot, Err: sgn.Err, Status: sgn.ConfirmationStatus})
    }
    return &Response{Items: items}, nil
}
