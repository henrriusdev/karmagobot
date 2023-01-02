package main

import (
	"database/sql"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"karmagot/internal/karma"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(os.Stderr, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("KARMABOT_API"))
	if err != nil {
		errorLog.Fatal(err)
		return
	}

	db, err := openDB(os.Getenv("web:julieta@/karmabot?parseTime=true"))
	if err != nil {
		errorLog.Fatal(err)
		return
	}
	infoLog.Println("Starting bot...")

	karmas := karma.KarmaModel{DB: db}
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		chat := update.Message.Chat.UserName
		lastUpdated, err := karmas.GetLastUpdated(int(update.Message.From.ID), chat)
		if err != nil {
			errorLog.Fatal(err)
			return
		}

		if update.Message != nil {

			// For +1 or -1
			if strings.Contains(update.Message.Text, "+1") && !checkGiveKarma(lastUpdated) {
				err = karmas.AddKarma(int(update.Message.From.ID), chat)
				if err != nil {
					errorLog.Fatal(err)
					return
				}

			} else if strings.Contains(update.Message.Text, "-1") && !checkGiveKarma(lastUpdated) {
				err = karmas.AddKarma(int(update.Message.From.ID), chat)
				if err != nil {
					errorLog.Fatal(err)
					return
				}
			} else {
				error := tgbotapi.NewMessage(update.Message.Chat.ID, "ERROR, You must to have to wait 1 minute to give karma.")
				if _, err := bot.Send(error); err != nil {
					errorLog.Fatal(err)
					return
				}
				return
			}

			karma, err := karmas.GetActualKarma(int(update.Message.From.ID), chat)
			if err != nil {
				return
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "@"+update.Message.From.UserName+" has now "+strconv.Itoa(karma)+" of karma")
			msg.ReplyToMessageID = update.Message.MessageID
			if _, err := bot.Send(msg); err != nil {
				errorLog.Fatal(err)
				return
			}

			if update.Message.IsCommand() {
				cmdText := update.Message.Command()
				switch cmdText {
				case "/karma":
					userKarma, err := karmas.GetActualKarma(int(update.Message.From.ID), chat)
					if err != nil {
						errorLog.Fatal(err)
						return
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "@"+update.Message.From.UserName+" has "+strconv.Itoa(userKarma)+" of karma.")
					msg.ReplyToMessageID = update.Message.MessageID
					if _, err := bot.Send(msg); err != nil {
						errorLog.Fatal(err)
						return
					}
					break
				case "/karmalove":
					users, err := karmas.GetKarmas(chat, true)
					if err != nil {
						errorLog.Fatal(err)
						return
					}

					usersString := "Most loved users of " + chat + "\n"
					for i, user := range users {
						usersString += fmt.Sprintf("%d. %s, %d karma.\n", i, update.Message.From.UserName, user.Count)
					}
					break
				case "/karmahate":
					users, err := karmas.GetKarmas(chat, false)
					if err != nil {
						errorLog.Fatal(err)
						return
					}

					usersString := "Most hated users of " + chat + "\n"
					for i, user := range users {
						usersString += fmt.Sprintf("%d. %s, %d karma.\n", i, update.Message.From.UserName, user.Count)
					}
				}
			}
		}
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func checkGiveKarma(lastKarmaGived time.Time) bool {
	return lastKarmaGived.Day() == time.Now().Day() &&
		lastKarmaGived.Hour() == time.Now().Hour() &&
		lastKarmaGived.Minute() == time.Now().Minute()
}
