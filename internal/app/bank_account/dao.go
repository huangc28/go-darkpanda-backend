package bankAccount

import (
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type BankAccountDAO struct {
	db db.Conn
}

type BankAccount struct {
	models.BankAccount
}

func NewBankAccountDAO(db db.Conn) *BankAccountDAO {
	return &BankAccountDAO{
		db: db,
	}
}

func BankAccountDAOServiceProvider(c container.Container) func() error {
	return func() error {
		c.Transient(func() contracts.BankAccountDAOer {
			return NewBankAccountDAO(db.GetDB())
		})

		return nil
	}
}

func (dao *BankAccountDAO) WithTx(tx db.Conn) contracts.BankAccountDAOer {
	dao.db = tx

	return dao
}

func (dao *BankAccountDAO) GetUserBankAccount(uuid string) (*models.BankAccount, error) {
	query := `
		SELECT 
			ba.bank_name,
			ba.branch,
			ba.account_number,
			ba.verify_status 
		FROM users u 
		LEFT JOIN bank_accounts ba ON u.id=ba.user_id 
		WHERE u.uuid = $1;
	`

	bank := &models.BankAccount{}

	if err := dao.db.QueryRow(query, uuid).Scan(&bank.BankName, &bank.AccountNumber, &bank.Branch, &bank.VerifyStatus); err != nil {
		return nil, err
	}

	return bank, nil
}

func (dao *BankAccountDAO) CheckHasBankAccountByUUID(uuid string) (bool, error) {
	query := `
	SELECT EXISTS(
		SELECT 1
		FROM bank_accounts ba 
		INNER JOIN users u on ba.user_id=u.id
		WHERE u.uuid =$1
	) AS "exists"
	`
	var exists bool

	if err := dao.db.QueryRow(query, uuid).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (dao *BankAccountDAO) PatchBankAccount(uuid string, params contracts.PatchBankAccountParams) error {
	query := `
		UPDATE bank_accounts 
		SET bank_name = COALESCE($1, bank_name), 
			branch = COALESCE($2, branch), 
			account_number = COALESCE($3, account_number), 
			verify_status = COALESCE($4, verify_status)
		WHERE user_id=$5
	`

	_, err := dao.db.Exec(
		query,
		params.BankName,
		params.Branch,
		params.AccountNumber,
		params.VerifyStatus,
		params.UserID,
	)

	if err != nil {
		return err
	}

	return err
}

func (dao *BankAccountDAO) InsertBankAccount(params contracts.InsertBankAccountParams) error {
	query := `
		INSERT INTO bank_accounts (
			user_id, 
			bank_name, 
			branch, 
			account_number, 
			verify_status, 
			created_at, 
			updated_at, 
			deleted_at
		)
		VALUES($1, $2, $3, $4, $5, now(), CURRENT_TIMESTAMP, null);
	`

	_, err := dao.db.Exec(
		query,
		params.UserID,
		params.BankName,
		params.Branch,
		params.AccountNumber,
		params.VerifyStatus,
	)

	if err != nil {
		return err
	}

	return err
}
