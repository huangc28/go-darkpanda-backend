package block

import (
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type BlockDAO struct {
	db db.Conn
}

type Block struct {
	models.BlockList
}

func NewBlockDAO(db db.Conn) *BlockDAO {
	return &BlockDAO{
		db: db,
	}
}

func BankAccountDAOServiceProvider(c container.Container) func() error {
	return func() error {
		c.Transient(func() contracts.BlockDAOer {
			return NewBlockDAO(db.GetDB())
		})

		return nil
	}
}

func (dao *BlockDAO) WithTx(tx db.Conn) contracts.BlockDAOer {
	dao.db = tx

	return dao
}

func (dao *BlockDAO) GetBlockedUsers(uuid string) ([]models.User, error) {
	query := `
		SELECT 
			u2.uuid, 
			u2.username, 
			u2.avatar_url
		FROM 
			block_list bl 
		INNER JOIN users u ON bl.user_id = u.id 
		LEFT JOIN users u2 ON bl.blocked_user_id = u2.id
		WHERE u.uuid = $1;
	`

	rows, err := dao.db.Queryx(
		query,
		uuid,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]models.User, 0)

	for rows.Next() {
		var m models.User

		if err := rows.StructScan(&m); err != nil {
			return nil, err
		}

		users = append(users, m)
	}

	return users, nil
}

type BlockUserParams struct {
	BlockerId int
	BlockeeId int
}

func (dao *BlockDAO) BlockUser(p BlockUserParams) error {
	query := `
INSERT INTO block_list(
	user_id,
	blocked_user_id,
	deleted_at
) 
VALUES ($1, $2, current_timestamp) 
ON CONFLICT (user_id, blocked_user_id) DO UPDATE
SET 
	user_id = $1,
	blocked_user_id = $2,
	deleted_at = current_timestamp;
	`

	if _, err := dao.db.Exec(
		query,
		p.BlockerId,
		p.BlockeeId,
	); err != nil {
		return err
	}

	return nil
}

type UnblockParams struct {
	BlockerUuid string
	BlockeeUuid string
}

func (dao *BlockDAO) Unblock(p UnblockParams) error {
	query := `
WITH blocker_info AS (
	SELECT 
		id
	FROM 
		users 
	WHERE 
		uuid = $1  
), blockee_info AS (
	SELECT 
		id
	FROM 
		users 
	WHERE 
		uuid = $2  
)  
UPDATE 
	block_list
SET 
	deleted_at = null
WHERE
	user_id IN (
		SELECT id FROM blocker_info
	) AND	
	blocked_user_id IN (
		SELECT id FROM blockee_info
	);
	`
	_, err := dao.db.Exec(
		query,
		p.BlockerUuid,
		p.BlockeeUuid,
	)

	return err
}

type HasBlockedByUserParams struct {
	BlockerUuid string
	BlockeeUuid string
}

func (dao *BlockDAO) HasBlockedByUser(p HasBlockedByUserParams) (bool, error) {
	query := `
WITH blocker_info AS (
	SELECT 
		id
	FROM 
		users 
	WHERE 
		uuid = $1  
), blockee_info AS (
	SELECT 
		id
	FROM 
		users 
	WHERE 
		uuid = $2  
)  
SELECT EXISTS (
	SELECT 
		1 
	FROM 
		block_list
	WHERE
		user_id IN (
			SELECT id FROM blocker_info
		) AND
		blocked_user_id IN (
			SELECT id FROM blockee_info			
		) AND deleted_at IS NOT NULL	
);
	`
	var hasBlocked bool

	if err := dao.db.QueryRowx(
		query,
		p.BlockerUuid,
		p.BlockeeUuid,
	).Scan(&hasBlocked); err != nil {
		return false, err
	}

	return hasBlocked, nil
}
