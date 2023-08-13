package store

import (
	"database/sql"
)

type Novel struct {
	ID             int
	Title          string
	Author         string
	LibraryCardUrl string
	PostStatus     int
}

const (
	NotPosted           = 0
	Posted              = 1
	FileSizeOver        = 2
	CopyrightSurvival   = 3
	CloudNotGetNovelUrl = 4
)

func Insert(db *sql.DB, n Novel) error {
	// テーブル作成
	_, err := db.Exec(
		`CREATE TABLE IF NOT EXISTS "NOVEL" ("ID" INTEGER PRIMARY KEY AUTOINCREMENT, "TITLE" VARCHAR(255), "AUTHOR" VARCHAR(255), "LIBRARYCARDURL" VARCHAR(255), "POSTSTATUS" INTEGER)`,
	)
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`INSERT INTO NOVEL(TITLE,AUTHOR,LIBRARYCARDURL,POSTSTATUS) VALUES(?,?,?,?)`,
		n.Title,
		n.Author,
		n.LibraryCardUrl,
		NotPosted,
	)
	if err != nil {
		return err
	}
	return nil
}

//投稿済みに更新する
func UpdatePostStatus(db *sql.DB, id, postStatus int) error {
	_, err := db.Exec(
		`UPDATE NOVEL SET POSTSTATUS = ? WHERE id = ?`,
		postStatus,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

//未投稿の小説一覧を取得する
func GetNotPostedNovels(db *sql.DB) ([]Novel, error) {
	ns := []Novel{}
	rows, err := db.Query(`SELECT ID, TITLE, AUTHOR, LIBRARYCARDURL, POSTSTATUS FROM NOVEL WHERE POSTSTATUS = ?`, NotPosted)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		n := Novel{}
		err := rows.Scan(&n.ID, &n.Title, &n.Author, &n.LibraryCardUrl, &n.PostStatus)
		if err != nil {
			return nil, err
		}
		ns = append(ns, n)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return ns, nil
}
