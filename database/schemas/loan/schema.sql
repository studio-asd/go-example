-- drop tables.
DROP TABLE IF EXISTS loans;
DROP TABLE IF EXISTS loan_instalments;
DROP TABLE IF EXISTS loan_invoices;
DROP TABLE IF EXISTS loan_bills;
DROP TABLE IF EXISTS loan_payments;

-- types.

DROP TYPE IF EXISTS loan_status;
CREATE TYPE loan_status AS ENUM('active', 'inactive', 'paid', 'stale');

DROP TYPE IF EXISTS instalment_status;
CREATE TYPE instalment_status AS ENUM('active', 'closed', 'paid');

DROP TYPE IF EXISTS instalment_billable_every;
CREATE TYPE instalment_billable_every AS ENUM('yearly', 'monthly', 'biweekly', 'weekly', 'daily');

DROP TYPE IF EXISTS invoice_status;
CREATE TYPE invoice_status AS ENUM('active', 'paid', 'cancelled')

DROP TYPE IF EXISTS invoice_type
CREATE TYPE invoice_type AS ENUM('instalment', 'loan')

DROP TYPE IF EXISTS bill_type;
CREATE TYPE bill_type AS ENUM('invoice', 'closing');

DROP TYPE IF EXISTS bill_status;
CREATE TYPE bill_status AS ENUM('active', 'paid', 'cancelled');

-- tables and index.

-- loans is the list of loans created for the client.
CREATE TABLE IF NOT EXISTS loans(
	"loan_id" varchar PRIMARY KEY,
	"loan_status" loan_status NOT NULL,
	"client_id" varchar NOT NULL,
	-- loan_amount is the amount of loan.
	"loan_amount" numeric NOT NULL,
	-- total_interest is the total of interest for the loan.
	"total_interest" numeric NOT NULL,
	"currency_id" int NOT NULL,
	-- loan_start_date is the time when the loan is started.
	"loan_start_date" timestamptz NOT NULL,
	-- loan_end_date is the time when the loan is ended.
	"loan_end_date" timestamptz NOT NULL,
	-- loan_paid_by_invoice_id record which invoice fully paid the loan.
	"loan_paid_by_invoice_id" varchar NOT NULL,
	-- loan_paid_by_bill_id record which bill fully paid the loan. We might be able to look this by looking by the invoice, but it
	-- also doesn't hurt to record this here.
	"loan_paid_by_bill_id" varchar NOT NULL,
	-- finished_date is the time when the loan is finished, this means the client has pay for all the instalments.
	-- A loan can be finished before or after of the end date.
	"finished_date" timestamptz NOT NULL,
	-- idempotency_key protects double creation of a loan. As loan is usually triggered from an approval process, there should
	-- be a unique identifier for it.
	"idempotency_key" varchar NOT NULL,
	"created_at" timestamptz NOT NULL,
	"updated_at" timestamptz,
	UNIQUE(idempotency_key)
);

-- loan_instalment is the detail of instalment and interest for a given instalment. A loan can have multiple stage of instalment based
-- on how the loan is structured.
--
-- The idea to store the billable_dates, billable_invoice and payment_dates is to save space and computation complexity to retrieve the data
-- needed for several things:
--
-- 1. We need to issue a billable_invoice for the client so the client is able to pay the invoice. This means we only need to look
--    at the 'billable_dates' and check whether it contains the current_date or in range of invoice creation. But there's a catch, what if
--    the invoice is already created? This is why we store 'billable_invoice' here, we will not issue any invoice anymore if the length of
--    the 'billable_dates' == 'billable_invoice'.
--
--    Okay, but loan can be not as that straightforward, what if a given user want to pay upfront or change their instalment to a longer date?
--    This will be answered by the 'instalment_paid_by_invoice' and 'instalment_status'. We will both know what invoice is used to change the status
--    of an instalment to paid, and we wknow whether an instalment is still active or not.
--    In case of the change of instalment contract, we need to void the current instalment and create a new one under the same loan.
--
-- 2. We need to check whether a given client have paid their invoice or not by using 'paid_invoice'. If the number of 'paid_invoice' is less than
--    the 'billable_invoice', then we can take some actions from there.
CREATE TABLE IF NOT EXISTS loan_instalments(
	"instalment_id" varchar PRIMARY KEY,
	"loan_id" varchar NOT NULL,
	"client_id" varchar NOT NULL,
	-- loan_amount is the total amount of money loaned to the client.
	"loan_amount" numeric NOT NULL,
	-- interest_percentage is the total percentage interest for the specific instalment_id.
	"interest_percentage" int NOT NULL,
	-- instalment_amount is the total amount of instalment for the interest_id.
	"instalment_amount" numeric NOT NULL,
	-- billable_amount is the amount of money the client need to pay for every invoice.
	"billable_amount" numeric NOT NULL,
	-- start_date is the start_date of the instalment.
	"instalment_start_date" timestamptz NOT NULL,
	-- end_date is the end_date of the instalment, if the number of instalment is one(1), then the end date will be
	-- the same with the end date of a loan.
	"instalment_end_date" timestamptz NOT NULL,
	-- total_billable_amount is the total of bilable invoice for the instalment.
	"total_billable_amount" numeric NOT NULL,
	-- total_paid_amount is the total of paid amount for the instalment.
	"total_paid_amount" numeric NOT NULL,
	-- billable_dates is the time to create a invoice bill to the customer.
	"billable_dates" date[] NOT NULL,
	-- billable_invoices is the list of invoice issued for the billable_dates.
	"billable_invoices" varchar[],
	-- billable_paid_invoice is the list of invoice that have been being paid by the client.
	"billable_paid_invoices" date[],
	-- instalment_paid_by_invoice mark which invoice is used to fully paid the instalment. The invoice used to pay the instalment might not
	-- be from the billable_invoice that automatically generated. For example, if a given client want to fully paid their loan, we can issue
	-- a new invoice to pay all the instalments.
	"instalment_paid_by_invoice" varchar,
	-- finished_at is the time when the instalment is fully paid.
	"finished_at" timestamptz,
	"created_at" timestamptz NOT NULL,
	"updated_at" timestamptz
);

DROP INDEX IF EXISTS idx_loan_instalment_billable_dates;
CREATE INDEX IF NOT EXISTS idx_loan_instalment_billable_dates ON loan_instalment(billable_dates);
DROP INDEX IF EXISTS idx_loan_instalment_billable_invoices;
CREATE INDEX IF NOT EXISTS idx_loan_instalment_billable_invoices ON loan_instalment(billable_invoices);

-- loan_invoice records all invoice for a given loan. The invoice can be used for instalment or non-instalment charge for the loan.
-- For example, we might issue an issue for administration/overdue payment separately from the instalment. While its being issued as
-- a separate invoice, but it can still be billed together.
CREATE TABLE IF NOT EXISTS loan_invoice(
	"invoice_id" varchar PRIMARY KEY,
	"invoice_type" invoice_type NOT NULL,
	-- instalment_id refers the instalment_id if the invoice_type is 'instalment'.
	"instalment_id" varchar,
	"loan_id" varchar NOT NULL,
	"user_id" varchar NOT NULL,
	-- paid_by_bill_id is an identifier to flag that the invoice have been paid via a billing_id.
	"paid_by_bill_id" varchar NOT NULL,
	"amount" numeric NOT NULL,
	"invoice_status" invoice_status NOT NULL,
	"created_at" timestamptz NOT NULL,
	"updated_at" timestamptz
)

-- loan_bills provide an aggregated billing of invoices in an loan. A bill is created when an invoice is issued for the loan.
-- If somehow, the bill is not get paid before the next invoice creation, then we will append the next invoice into the
-- current bill.
--
-- A client might pay less than the billed amount.
CREATE TABLE IF NOT EXISTS loan_biils(
	"bill_id" varchar PRIMARY KEY,
	"bill_type" bill_type NOT NULL,
	"bill_status" bill_status NOT NULL,
	-- previous_bill_id is used when generating a new billing for the same loan. This to ensure we are not processing the same billing
	-- twice and ensure we are processing bills in sequence.
	"previous_bill_id" varchar NOT NULL,
	"loan_id" varchar NOT NULL,
	"user_id" varchar NOT NULL,
	"total_amount" numeric NOT NULL,
	"total_paid" numeric NOT NULL,
	-- invoices is the list of invoice inside the bills. The question might be, what if some invoices are being paid by another bill?
	-- Because there can only be one active bill per loan, the race between bills cannot happen.
	"invoices" varchar[] NOT NULL,
	-- payments stores all payment_id for a given bill. A bill can be paid once or many times depends on how the
	-- client pay the bill.
	"payments" varchar[],
	-- payment_due_date is the due date of a payment for a bill. The date can be extended.
	"payment_due_date" timestamptz NOT NULL,
	-- finished_at records the time the bill transitioned into the final status.
	"finished_at" timestamptz,
	"crearted_at" timestamptz NOT NULL,
	"updated_at" timestamptz,
	UNIQUE(loan_id, previous_bill_id)
);

DROP INDEX IF EXISTS idx_loan_bills_payment_due_date;
CREATE INDEX IF NOT EXISTS idx_loan_bills_payment_due_date ON loan_bills(payment_due_date);

CREATE TABLE IF NOT EXISTS loan_payments(
	"payment_id" varchar PRIMARY KEY,
	-- idempotency_key ensures the payment transactions are unique.
	"idempotency_key" varchar NOT NULL,
	"loan_id" varchar NOT NULL,
	"user_id" varchar NOT NULL,
	-- bill_id ensure a payment is targeted to a certain billing.
	"bill_id" varchar NOT NULL,
	"amount" numeric NOT NULL,
	"created_at" timestamptz NOT NULL,
	UNIQUE(idempotency_key)
);
