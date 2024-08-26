-- name: CreateLoan :one
INSERT INTO loans(
	loan_id, -- 1
	loan_status, -- 2
	client_id, -- 3
	loan_amount, -- 4
	total_interest, -- 5
	currency_id, -- 6
	loan_start_date, -- 7
	loan_end_date, -- 8
	idempotency_key, --9
	created_at -- 10
) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);

-- name: CreateInstalment :one
INSERT INTO loan_instalments(
	instalment_id, -- 1
	loan_id, -- 2
	client_id, -- 3
	interest_percentage, -- 4
	instalment_amount, -- 5
	billable_amount, -- 6
	instalment_star_date, -- 7
	instalment_end_date, -- 8
	total_billable_amount, -- 9
	total_paid_amount, -- 10
	billable_dates, -- 11
	created_at -- 12
) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12);

-- name: GetBillableInstalmentByDate :many
SELECT *
FROM loan_instalment
WHERE $1 = ANY(billable_dates)
AND NOT ($2 = ANY(billable_invoices));