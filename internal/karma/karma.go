package karma

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type Karma struct {
	User        int
	Count       int
	LastUpdated time.Time
}

type KarmaModel struct {
	DB *sql.DB
}

// CREATE & INSERT METHODS

func (m *KarmaModel) CreateTable(channel string, users ...int) error {
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s`(user_id INT NOT NULL PRIMARY KEY, karma INT NOT NULL, last_updated DATETIME NOT NULL)", channel)
	_, err := m.DB.Exec(query)
	if err != nil {
		return err
	}

	query = fmt.Sprintf("INSERT IGNORE INTO `%s`(user_id, karma, last_updated) VALUES (?, ?, ?)")
	for user := range users {
		_, err = m.DB.Exec(query, user, 0, time.Now())
		if err != nil {
			return err
		}
	}

	fmt.Println("Created table for: ", channel, "group")

	return nil
}

// GET methods

func (m *KarmaModel) GetKarmas(channel string, top bool) ([]*Karma, error) {
	var query string
	if top {
		query = fmt.Sprintf("SELECT user, karma FROM `%s` ORDER BY ASC TOP 10", channel)
	} else {
		query = fmt.Sprintf("SELECT user, karma FROM `%s` ORDER BY DESC TOP 10", channel)
	}

	rows, err := m.DB.Query(query)
	defer rows.Close()

	var karmas []*Karma

	for rows.Next() {
		var karma *Karma
		if err = rows.Scan(&karma.User, &karma.Count); err != nil {
			if err == sql.ErrNoRows {
				return nil, err
			}

			return nil, err
		}

		karmas = append(karmas, karma)
	}
	return karmas, nil
}

func (m *KarmaModel) GetActualKarma(userId int, channel string) (int, error) {
	var user Karma
	query := fmt.Sprintf("SELECT karma FROM `%s` WHERE user_id = ?", channel)

	if err := m.DB.QueryRow(query, userId).Scan(&user.Count); err != nil {
		if err == sql.ErrNoRows {
			return 0, err
		}
		return 0, err
	}

	return user.Count, nil
}

func (m *KarmaModel) GetLastUpdated(userId int, channel string) (time.Time, error) {
	var user Karma
	query := fmt.Sprintf("SELECT last_updated FROM `%s` WHERE user_id = ?", channel)

	if err := m.DB.QueryRow(query, userId).Scan(&user.LastUpdated); err != nil {
		if err == sql.ErrNoRows {
			return user.LastUpdated, err
		}
		return user.LastUpdated, err
	}

	return user.LastUpdated, nil
}

// UPDATE METHODS

func (m *KarmaModel) AddKarma(userId int, channel string) error {
	query := fmt.Sprintf("UPDATE `%s` SET Karma = ? WHERE user_id = ?", channel)

	karma, err := m.GetActualKarma(userId, channel)
	if err != nil {
		return err
	}
	karma++
	_, err = m.DB.Exec(query, userId, karma, time.Now())
	if err != nil {
		return err
	}

	err = m.updateLastKarma(time.Now(), channel, userId)
	if err != nil {
		return err
	}

	return nil
}

func (m *KarmaModel) SubstractKarma(userId int, channel string) error {
	query := fmt.Sprintf("UPDATE `%s` SET Karma = ? WHERE user_id = ?", channel)

	karma, err := m.GetActualKarma(userId, channel)
	if err != nil {
		return err
	}
	karma--
	_, err = m.DB.Exec(query, userId, karma, time.Now())
	if err != nil {
		return err
	}

	err = m.updateLastKarma(time.Now(), channel, userId)
	if err != nil {
		return err
	}

	return nil
}

func (m *KarmaModel) updateLastKarma(date time.Time, channel string, userId int) error {
	query := fmt.Sprintf("UPDATE `%s` SET last_updated = ? WHERE user_id = ?", channel)
	_, err := m.DB.Exec(query, userId, date, time.Now())
	if err != nil {
		return err
	}

	return nil
}
