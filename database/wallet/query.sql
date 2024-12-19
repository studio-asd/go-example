-- name: CreateWallet :exec
INSERT INTO wallet_accounts(
	wallet_id,
	ledger_account_id,
	user_id,
	wallet_status,
	wallet_type,
	created_at
) VALUES($1,$2,$3,$4,$5);