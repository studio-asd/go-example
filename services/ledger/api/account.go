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
	ledgerv1 "github.com/albertwidi/go-example/proto/api/ledger/v1"
	"github.com/albertwidi/go-example/services/ledger"
	ledgerpg "github.com/albertwidi/go-example/services/ledger/internal/postgres"
)

// CreateAccounts API allows the client to create ledger accounts. It is possible to associate an account_id as the parent_id when
// creating the account. The API don't allow inactive account to be referenced as parent_id.
func (a *API) CreateAccounts(ctx context.Context, request *ledgerv1.CreateLedgerAccountsRequest, fn func(context.Context, *postgres.Postgres, []ledger.AccountInfo) error) (*ledgerv1.CreateLedgerAccountsResponse, error) {
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

	// accountInfo is the information we kept for the callback function to allow them to consume the informations.
	var accountInfo = make([]ledger.AccountInfo, len(request.Accounts))
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
		accountInfo[idx] = ledger.AccountInfo{
			AccountID:       accID,
			ParentAccountID: acc.ParentAccountId,
			AllowNegative:   acc.AllowNegative,
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
	if fn == nil {
		// Nil out the account info as soon as possible because we don't need this in non-callback case.
		accountInfo = nil
		return resp, a.queries.CreateLedgerAccounts(ctx, createReqs...)
	}

	err := a.queries.Postgres().Transact(ctx, sql.LevelReadCommitted, func(ctx context.Context, p *postgres.Postgres) error {
		if err := ledgerpg.New(p).CreateLedgerAccounts(ctx, createReqs...); err != nil {
			return err
		}
		_, err := await.Do(ctx, time.Second*3, accountInfo, func(ctx context.Context, info []ledger.AccountInfo) (any, error) {
			return nil, fn(ctx, p, info)
		})
		return err
	})
	if err != nil {
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
