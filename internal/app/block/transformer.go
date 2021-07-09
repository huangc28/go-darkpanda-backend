package block

import "github.com/huangc28/go-darkpanda-backend/internal/app/models"

type BlockTransform struct{}

func NewTransform() *BlockTransform {
	return &BlockTransform{}
}

type TransformedBlock struct {
	Uuid      string `form:"uuid" json:"uuid"`
	Username  string `form:"username" json:"username"`
	AvatarUrl string `form:"avatar_url" json:"avatar_url"`
}

type TransformedBlocks struct {
	Blocklist []TransformedBlock `json:"blocked_users"`
}

func (ba *BlockTransform) TransformBlockedUser(users []models.User) *TransformedBlocks {
	blocks := make([]TransformedBlock, 0)

	for _, user := range users {
		block := TransformedBlock{
			Uuid:      user.Uuid,
			Username:  user.Username,
			AvatarUrl: user.AvatarUrl.String,
		}

		blocks = append(blocks, block)
	}

	return &TransformedBlocks{blocks}
}
