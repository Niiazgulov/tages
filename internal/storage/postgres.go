package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type ImageDB interface {
	SaveNewInfo(imageInfo ImagesInfo) error
	UpdateInfo(imageInfo ImagesInfo) (string, error)
	GetAllInfo() ([]ImagesInfo, error)
	Close()
}

type DataBase struct {
	DB *sql.DB
}

func NewDB(dbPath string) (ImageDB, error) {
	db, err := sql.Open("pgx", dbPath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS images (
			id SERIAL PRIMARY KEY,
			filename VARCHAR UNIQUE,
			created_at VARCHAR, 
			changed_at VARCHAR,
			image_id VARCHAR)
		`)
	if err != nil {
		return nil, fmt.Errorf("unable to CREATE TABLE in DB: %w", err)
	}

	return &DataBase{DB: db}, nil
}

func (d *DataBase) SaveNewInfo(imageInfo ImagesInfo) error {
	query := `INSERT INTO images (image_id, filename, created_at, changed_at) VALUES ($1, $2, $3, $4)`
	_, err := d.DB.Exec(query, imageInfo.ImageId, imageInfo.Filename, imageInfo.CreatedAt, imageInfo.ChangedAt)
	if err != nil && strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
		query := `UPDATE images SET changed_at = $1 WHERE filename = $2`
		_, err := d.DB.Exec(query, imageInfo.ChangedAt, imageInfo.Filename)
		if err != nil {
			return fmt.Errorf("[Image DB] Error while updating new image info (SaveNewInfo): %w", err)
		}

		return nil
	}
	if err != nil {
		return fmt.Errorf("[Image DB] Error while SAVING new image info: %w", err)
	}

	return nil
}

func (d *DataBase) UpdateInfo(imageInfo ImagesInfo) (string, error) {
	query := `UPDATE images SET changed_at = $1 WHERE filename = $2 RETURNING image_id`
	row := d.DB.QueryRow(query, imageInfo.ChangedAt, imageInfo.Filename)
	var imageID string
	if err := row.Scan(&imageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("[Image DB] key not found while updating DB: %w", err)
		}
		return "", fmt.Errorf("[Image DB] I unable to Scan imageID from DB: %w", err)
	}

	return imageID, nil
}

func (d *DataBase) GetAllInfo() ([]ImagesInfo, error) {
	query := `SELECT filename, created_at, changed_at FROM images`
	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("unable to return records from DB: %w", err)
	}
	defer rows.Close()

	records := []ImagesInfo{}
	for rows.Next() {
		record := ImagesInfo{}
		err = rows.Scan(&record.Filename, &record.CreatedAt, &record.ChangedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (d DataBase) Close() {
	d.DB.Close()
}
