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

func (dao *BlockDAO) GetUserBlock(uuid string) ([]contracts.GetUserBlockListParams, error) {
	query := `
		SELECT bl.id, u2.id AS user_id, u2.username, u2.avatar_url
		FROM block_list bl 
		INNER JOIN users u ON bl.user_id = u.id 
		LEFT JOIN users u2 ON bl.blocked_user_id = u2.id
		WHERE u.uuid = $1;
	`

	rows, err := dao.db.Query(
		query,
		uuid,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	blocks := make([]contracts.GetUserBlockListParams, 0)
	for rows.Next() {
		var block contracts.GetUserBlockListParams

		if err := rows.Scan(&block.ID, &block.UserId, &block.Username, &block.AvatarUrl); err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (dao *BlockDAO) InsertUserBlock(params contracts.InsertUserBlockListParams) error {
	query := `
		INSERT INTO block_list(
			user_id,
			blocked_user_id
		) VALUES ($1, $2);
	`

	if _, err := dao.db.Exec(
		query,
		params.UserId,
		params.BlockedUserId,
	); err != nil {
		return err
	}

	return nil
}

func (dao *BlockDAO) DeleteUserBlock(blockId string) error {
	query := `
		DELETE FROM block_list 
		WHERE id=$1
	`

	_, err := dao.db.Exec(
		query,
		blockId,
	)

	if err != nil {
		return err
	}

	return err
}
