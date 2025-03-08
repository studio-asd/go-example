package api

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/studio-asd/pkg/postgres"
	"github.com/studio-asd/pkg/srun"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/studio-asd/go-example/internal/await"
	"github.com/studio-asd/go-example/internal/protovalidate"
	ledgerv1 "github.com/studio-asd/go-example/proto/api/ledger/v1"
	"github.com/studio-asd/go-example/services/ledger"
	internal "github.com/studio-asd/go-example/services/ledger/internal"
	ledgerpg "github.com/studio-asd/go-example/services/ledger/internal/postgres"
)

var (
	validator *protovalidate.Validator
	_         srun.ServiceInitAware = (*API)(nil)
)

func init() {
	var err error
	validator, err = protovalidate.New(
		protovalidate.WithFailFast(),
		protovalidate.WithMessages(
			&ledgerv1.TransactRequest{},
			&ledgerv1.CreateLedgerAccountsRequest_Account{},
			&ledgerv1.GetAccountsBalanceRequest{},
		),
	)
	if err != nil {
		panic(err)
	}
}

type API struct {
	queries *ledgerpg.Queries
	logger  *slog.Logger
}

func New(pg *postgres.Postgres) *API {
	return &API{
		queries: ledgerpg.New(pg),
	}
}

func (a *API) Name() string {
	return "ledger_api"
}

func (a *API) Init(ctx srun.Context) error {
	a.logger = ctx.Logger
	return nil
}

// GRPC returns the grpc api implementation of the ledger api.
func (a *API) GRPC() *GRPC {
	return newGRPC(a)
}

// Transact moves money from accounts to accounts within the transaction scope.
func (a *API) Transact(ctx context.Context, req *ledgerv1.TransactRequest, fn func(context.Context, *postgres.Postgres, ledger.MovementInfo) error) (*ledgerv1.TransactResponse, error) {
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	accounts := make([]string, len(req.GetMovementEntries())*2)
	entries := req.GetMovementEntries()
	for idx, entry := range entries {
		accounts[idx] = entry.FromAccountId
		accounts[idx+1] = entry.ToAccountId
	}

	accountsBalance, err := a.queries.GetAccountsBalanceMappedByAccID(ctx, accounts...)
	if err != nil {
		return nil, err
	}
	// Create a new UUID_V7 for movement_id.
	uuidv7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	ledgerEntries, err := createLedgerEntries(uuidv7.String(), req.GetIdempotencyKey(), accountsBalance, entries...)
	if err != nil {
		return nil, err
	}

	var result internal.MovementResult
	// If the additional function scope is not nil, then we should invoke the function inside a time-bounded
	// goroutine as we don't know how much time the function will spent. So we need to ensure the function runs
	// inside the Transact SLA.
	if fn == nil {
		result, err = a.queries.Move(ctx, ledgerEntries)
		if err != nil {
			return nil, err
		}
	} else {
		err = a.queries.WithTransact(ctx, sql.LevelReadCommitted, func(ctx context.Context, q *ledgerpg.Queries) error {
			result, err = q.Move(ctx, ledgerEntries)
			if err != nil {
				return err
			}
			info := ledger.MovementInfo{
				MovementID: result.MovementID,
			}
			err := q.Do(ctx, func(ctx context.Context, pg *postgres.Postgres) error {
				_, err := await.Do(ctx, time.Second*3, info, func(ctx context.Context, info ledger.MovementInfo) (any, error) {
					return nil, fn(ctx, pg, info)
				})
				return err
			})
			return err
		})
		if err != nil {
			return nil, err
		}
	}

	// Construct the response. As the movement id and ledger ids are constructed beforehand, we only consruct the response
	// after we know all operations is a success to not wasting compute resource.
	response := &ledgerv1.TransactResponse{
		MovementId:     ledgerEntries.MovementID,
		TransactTime:   timestamppb.New(result.Time),
		LedgerEntries:  make([]*ledgerv1.TransactResponse_LedgerEntry, len(ledgerEntries.LedgerEntries)),
		EndingBalances: make([]*ledgerv1.TransactResponse_Balance, len(result.Balances)),
	}
	for idx, entry := range ledgerEntries.LedgerEntries {
		response.LedgerEntries[idx] = &ledgerv1.TransactResponse_LedgerEntry{
			LedgerId: entry.LedgerID,
			ClientId: entry.ClientID,
			// Its okay to cast the movement seuqnce to int32 as it won't overflow as we don't allow huge entries.
			MovementSequence: int32(entry.MovementSequence),
		}
	}
	counter := 0
	for _, balance := range result.Balances {
		response.EndingBalances[counter] = &ledgerv1.TransactResponse_Balance{
			AccountId:          balance.AccountID,
			LedgerId:           balance.NextLedgerID,
			NewBalance:         balance.NewBalance.String(),
			PreviousBalance:    balance.PreviousBalance.String(),
			PreviousLedgerId:   balance.PreviousLedgerID,
			PreviousMovementId: balance.PreviousMovementID,
		}
		counter++
	}
	return response, nil
}
