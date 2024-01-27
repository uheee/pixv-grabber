package manifest

type WorkModel struct {
	Id         uint64 `db:"id"`
	Title      string `db:"title"`
	IllustType int    `db:"illust_type"`
	Tags       []byte `db:"tags"`
	PageCount  int    `db:"page_count"`
	CreateDate int64  `db:"create_date"`
	UpdateDate int64  `db:"update_date"`
	MaskedDate int64  `db:"masked_date"`
}
