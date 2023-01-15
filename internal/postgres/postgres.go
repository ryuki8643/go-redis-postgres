package postgres

import (
	"database/sql"
	"fmt"
	"github.com/geolocket/batch_redis/internal/momentType"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"log"
	"os"
	"time"
)

var env = os.Getenv("ENV")

var (
	host     = "host.docker.internal"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "dev"
)

func dbOpen() (*sql.DB, error) {
	if env != "" {
		host = os.Getenv("DATABASE_HOST")
		user = os.Getenv("DATABASE_USER")
		password = os.Getenv("DATABASE_PASSWORD")
		dbname = "postgres"
	}
	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlConn)

	if err != nil {
		return nil, errors.WithStack(err)
	}
	return db, errors.WithStack(err)
}

func DeleteMessages() error {
	db, err := dbOpen()
	defer db.Close()
	if err != nil {
		return errors.WithStack(err)
	}
	tx, err := db.Begin()
	if err != nil {
		return errors.WithStack(err)
	}

	rows, err := tx.Query("select min(\"recentReadAt\"),\"groupId\" from \"ChatGroupUser\" group by \"groupId\"")
	if err != nil {
		return errors.WithStack(err)
	}
	for rows.Next() {
		var minReadAt string
		var groupId string
		rows.Scan(&minReadAt, &groupId)

		_, err = db.Query("delete from \"Message\" where \"groupId\" = $1 and \"sendAt\" < $2", groupId, minReadAt)
		if err != nil {
			return errors.WithStack(err)
		}
		log.Println("a", minReadAt, groupId)

	}
	return err
}

func UpdateMoment(userMoments map[string]momentType.RedisMomentBody) error {
	db, err := dbOpen()
	defer db.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	tx, err := db.Begin()
	if err != nil {
		return errors.WithStack(err)
	}
	for i, v := range userMoments {
		_, err = tx.Exec("update \"UserLatestMoment\" set (\"accuracy\", \"activity\", \"areaLandedAt\", \"battery\", \"heading\", \"isCharging\", \"isMoving\", \"lat\", \"lng\", \"movingSpeed\",\"updatedAt\") = ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) where \"userId\" = $12", v.Accuracy, v.Activity, v.AreaLandedAt, v.Battery, v.Heading, v.IsCharging, v.IsMoving, v.LatLng[0], v.LatLng[1], v.MovingSpeed, time.Now(), i)
		if err != nil {
			tx.Rollback()
			return errors.WithStack(err)
		}
	}

	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return err

}
