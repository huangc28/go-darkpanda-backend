package inquiry

import "database/sql"

type UserDaoer interface {
	CheckIsMaleByUuid(uuid string) (bool, error)
	CheckIsFemaleByUuid(uuid string) (bool, error)
}

type InquiryDAOer interface {
	CheckHasActiveInquiryByID(id int64) (bool, error)
}

type InquiryDAO struct {
	db *sql.DB
}

func NewInquiryDAO(db *sql.DB) InquiryDAOer {
	return &InquiryDAO{
		db: db,
	}
}

func (dao *InquiryDAO) CheckHasActiveInquiryByID(id int64) (bool, error) {
	sql := `
SELECT EXISTS(
	SELECT 1 FROM users
	LEFT JOIN service_inquiries as si ON si.inquirer_id = users.id
	WHERE users.id = $1
	AND inquiry_status='inquiring'
) as exists;
`
	var exists bool

	err := dao.db.QueryRow(sql, id).Scan(&exists)

	return exists, err
}
