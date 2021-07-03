package referral

import (
	"errors"
	"fmt"
	"log"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type ReferralCodeDAO struct {
	db db.Conn
}

func NewReferralCodeDAO(db db.Conn) *ReferralCodeDAO {
	return &ReferralCodeDAO{
		db: db,
	}
}

func (dao *ReferralCodeDAO) GetByRefCode(refCode string, fields []string) (*models.UserRefcode, error) {
	columnstr := db.ComposeFieldsSQLString(fields...)

	baseQuery := `
SELECT
	%s
FROM
	user_refcodes
WHERE
	ref_code = $1
	`

	query := fmt.Sprintf(baseQuery, columnstr)

	var refCodeModel models.UserRefcode

	if err := dao.db.QueryRowx(query, refCode).StructScan(&refCodeModel); err != nil {
		return nil, err
	}

	return &refCodeModel, nil
}

type UpdateReferralCodeParams struct {
	ID        *int64
	InviteeID *int64
	RefCode   *string
}

func (dao *ReferralCodeDAO) UpdateReferralCodeByID(params UpdateReferralCodeParams) error {
	if params.ID == nil {
		return errors.New("id is required to update referral code")
	}

	query := `
UPDATE user_refcodes SET
	invitee_id = COALESCE($1, invitee_id),
	ref_code = COALESCE($2, ref_code)
WHERE
	id = $3;
`

	err := dao.db.QueryRow(
		query,
		params.InviteeID,
		params.RefCode,
		params.ID,
	).Err()

	if err != nil {
		return err
	}

	return nil
}

func (dao *ReferralCodeDAO) GetUnoccupiedReferralCode(userUuid string) (*models.UserRefcode, error) {
	query := `
		SELECT 
			user_refcodes.*
		FROM
			user_refcodes		
		INNER JOIN users ON users.id = user_refcodes.invitor_id
		WHERE
			users.uuid = $1 AND		  
			user_refcodes.invitee_id IS NULL
		ORDER BY user_refcodes.created_at DESC 
		LIMIT 1;
	`

	var m models.UserRefcode

	if err := dao.db.QueryRowx(query, userUuid).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

type CreateReferralCodeParams struct {
	InvitorID int
	RefCode   int
}

func (dao *ReferralCodeDAO) CreateReferralCode(p CreateReferralCodeParams) (*models.UserRefcode, error) {
	log.Printf("DEBUG ref code %v %v", p.InvitorID, p.RefCode)

	query := `
	INSERT INTO user_refcodes (
		invitor_id,
		ref_code,
		ref_code_type
	) VALUES ($1, $2, $3)
	RETURNING *;
	`
	var m models.UserRefcode

	if err := dao.db.QueryRowx(
		query,
		p.InvitorID,
		p.RefCode,
		models.RefCodeTypeInvitor,
	).StructScan(&m); err != nil {
		return (*models.UserRefcode)(nil), err
	}

	return &m, nil
}
