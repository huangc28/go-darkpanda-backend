package block

import (
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
)

type BlockTransform struct{}

func NewTransform() *BlockTransform {
	return &BlockTransform{}
}

type TransformedBlock struct {
	ID        int    `form:"id" json:"id"`
	UserId    int    `form:"user_id" json:"user_id"`
	Username  string `form:"username" json:"username"`
	AvatarUrl string `form:"avatar_url" json:"avatar_url"`
}

type TransformedBlocks struct {
	Blocklist []TransformedBlock `json:"block"`
}

func (ba *BlockTransform) TransformBlock(block []contracts.GetUserBlockListParams) (*TransformedBlocks, error) {
	blocks := make([]TransformedBlock, 0)

	for _, bloc := range block {
		block := TransformedBlock{
			ID:        bloc.ID,
			UserId:    bloc.UserId,
			Username:  bloc.Username,
			AvatarUrl: bloc.AvatarUrl.String,
		}

		blocks = append(blocks, block)
	}

	return &TransformedBlocks{
		blocks,
	}, nil
}
