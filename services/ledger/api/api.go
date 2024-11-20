package api

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/albertwidi/pkg/postgres"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/albertwidi/go-example/internal/await"
	"github.com/albertwidi/go-example/internal/currency"
	"github.com/albertwidi/go-example/internal/protovalidate"
	ledgerv1 "github.com/albertwidi/go-example/proto/api/ledger/v1"
	"github.com/albertwidi/go-example/services/ledger"
	ledgerpg "github.com/albertwidi/go-example/services/ledger/internal/postgres"
)

var validator *protovalidate.Validator

func init() {
	var err error
	validator, err = protovalidate.New(
		protovalidate.WithFailFast(true),
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
}

func New(queries *ledgerpg.Queries) *API {
	return &API{
		queries: queries,
	}
}

// Transact moves money from accounts to accounts within the transaction scope.
func (a *API) Transact(ctx context.Context, req *ledgerv1.TransactRequest, fn func(context.Context, *postgres.Postgres) error) (*ledgerv1.TransactResponse, error) {
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
	ledgerEntries, err := createLedgerEntries(uuid.NewString(), req.GetIdempotencyKey(), accountsBalance, entries...)
	if err != nil {
		return nil, err
	}

	// If the additional function scope is not nil, then we should invoke the function inside a time-bounded
	// goroutine as we don't know how much time the function will spent. So we need to ensure the function runs
	// inside the Transact SLA.
	if fn != nil {
		timeoutSLA := time.Second * 3
		err = await.Do(ctx, timeoutSLA, func(ctx context.Context) error {
			return a.queries.WithTransact(ctx, sql.LevelReadCommitted, func(ctx context.Context, q *ledgerpg.Queries) error {
				if err := q.Move(ctx, ledgerEntries); err != nil {
					return err
				}
				if err := q.Do(ctx, fn); err != nil {
					return err
				}
				return nil
			})
		})
	} else {
		if err := a.queries.Move(ctx, ledgerEntries); err != nil {
			return nil, err
		}
	}
	// Construct the response. As the movement id and ledger ids are constructed beforehand, we only consruct the response
	// after we know all operations is a success to not wasting compute resource.
	response := &ledgerv1.TransactResponse{
		MovementId: ledgerEntries.MovementID,
	}
	for _, le := range ledgerEntries.LedgerEntries {
		response.LedgerIds = append(response.LedgerIds, le.LedgerID)
	}
	if err != nil {
		return nil, err
	}
	return response, nil
}

// CreateAccounts API allows the client to create ledger accounts. It is possible to associate an account_id as the parent_id when
// creating the account. The API don't allow inactive account to be referenced as parent_id.
func (a *API) CreateAccounts(ctx context.Context, request *ledgerv1.CreateLedgerAccountsRequest) (*ledgerv1.CreateLedgerAccountsResponse, error) {
	if err := validator.Validate(request); err != nil {
		return nil, err
	}

	createdAt := time.Now()
	// Define the timestamppb here because all of data will have the same timestamp of craeted at.
	createdAtPb := timestamppb.New(createdAt)
	resp := &ledgerv1.CreateLedgerAccountsResponse{
		Accounts: make([]*ledgerv1.CreateLedgerAccountsResponse_Account, len(request.Accounts)),
	}
	createReqs := make([]ledgerpg.CreateLedgerAccount, len(request.Accounts))

	// If the parent account id is not empty, then we need to check whether the parent has another parent or not.
	// This is because we are not allowing sub of sub-account to be created at the first place.
	var parentAccountIDs []string
	for idx, acc := range request.Accounts {
		if acc.ParentAccountId != "" {
			parentAccountIDs = append(parentAccountIDs, acc.ParentAccountId)
		}
		accID := uuid.NewString()
		cur, err := currency.Currencies.GetByID(acc.CurrencyId)
		if err != nil {
			return nil, err
		}
		// Create the request upfront so we don't have to loop all over again.
		createReqs[idx] = ledgerpg.CreateLedgerAccount{
			AccountID:       accID,
			ParentAccountID: acc.ParentAccountId,
			AllowNegative:   acc.AllowNegative,
			AccountStatus:   ledgerpg.AccountStatusActive,
			Currency:        cur,
			CreatedAt:       createdAt,
		}
		// Create the response upfront so we don't have to loop the accounts all over again.
		resp.Accounts[idx] = &ledgerv1.CreateLedgerAccountsResponse_Account{
			AccountId: accID,
			CreatedAt: createdAtPb,
		}
	}
	if len(parentAccountIDs) > 0 {
		accs, err := a.queries.GetAccounts(ctx, parentAccountIDs)
		if err != nil {
			return nil, err
		}
		for _, acc := range accs {
			if acc.ParentAccountID != "" {
				return nil, fmt.Errorf("%w: cannot use account %s as the parent account. The account is registered as a sub-account", ledger.ErrAccountHasParent, acc.AccountID)
			}
			if acc.AccountStatus == ledgerpg.AccountStatusInactive {
				return nil, fmt.Errorf("%w: cannot use account %s as the parent account. The account is inactive", ledger.ErrAccountInactive, acc.AccountID)
			}
		}
	}
	// Save the data to the database.
	if err := a.queries.CreateLedgerAccounts(ctx, createReqs...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *API) GetAccountsBalance(ctx context.Context, req *ledgerv1.GetAccountsBalanceRequest) (*ledgerv1.GetAccountsBalanceResponse, error) {
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	resp := &ledgerv1.GetAccountsBalanceResponse{
		Balances: make([]*ledgerv1.AccountBalance, len(req.GetAccountIds())),
	}
	balances, err := a.queries.GetAccountsBalance(ctx, req.AccountIds)
	if err != nil {
		return nil, err
	}
	for idx, balance := range balances {
		resp.Balances[idx] = &ledgerv1.AccountBalance{
			AccountId:      balance.AccountID,
			Balance:        balance.Balance.String(),
			AllowNegative:  balance.AllowNegative,
			LastMovementId: balance.LastMovementID,
			LastLedgerId:   balance.LastLedgerID,
			UpdatedAt:      timestamppb.New(balance.UpdatedAt.Time),
		}
	}
	return resp, nil
}
