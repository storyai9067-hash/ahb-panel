package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI
var ownerID int64
var cardNumber string
var keyboardButtons []string

func main() {
	token := "YOUR_BOT_TOKEN"
	var err error
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	ownerID = 123456789

	keyboardButtons = []string{
		"📦 خرید کانفیگ",
		"ℹ️ راهنما",
		"👤 پروفایل من",
		"📞 پشتیبانی",
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		if update.Message != nil {
			msg := update.Message
			userID := msg.From.ID
			text := msg.Text

			switch text {
			case "📦 خرید کانفیگ":
				sendMessage(userID, "💳 برای خرید کانفیگ، مبلغ را به شماره کارت زیر واریز کنید:\n\n`"+cardNumber+"`\n\nبعد از واریز، رسید رو بفرستید.", true)
				continue

			case "ℹ️ راهنما":
				sendMessage(userID, "🤖 راهنمای ربات ahb panel:\n\n1️⃣ روی دکمه خرید کلیک کنید.\n2️⃣ به شماره کارت واریز کنید.\n3️⃣ رسید رو بفرستید تا کانفیگ دریافت کنید.\n4️⃣ برای اطلاعات بیشتر با پشتیبانی تماس بگیرید.", false)
				continue

			case "👤 پروفایل من":
				sendMessage(userID, "👤 پروفایل شما:\n\n📅 تاریخ عضویت: "+time.Now().Format("2006-01-02"), false)
				continue

			case "📞 پشتیبانی":
				supportKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL("📩 تماس با پشتیبانی", "https://t.me/trah90"),
					),
				)
				msg := tgbotapi.NewMessage(userID, "📞 برای ارتباط با پشتیبانی، روی دکمه زیر کلیک کن:")
				msg.ReplyMarkup = supportKeyboard
				bot.Send(msg)
				continue
			}

			if userID == ownerID {
				switch text {
				case "⚙️ تنظیمات":
					sendAdminPanel(userID)
					continue

				case "💳 تغییر شماره کارت":
					sendMessage(userID, "💳 شماره کارت جدید رو به صورت عددی وارد کن:", false)
					continue

				case "➕ افزودن دکمه":
					sendMessage(userID, "➕ اسم دکمه جدید رو وارد کن:", false)
					continue

				case "🗑️ حذف دکمه":
					if len(keyboardButtons) == 0 {
						sendMessage(userID, "⚠️ هیچ دکمه‌ای برای حذف وجود نداره.", false)
						continue
					}
					var list string
					for i, btn := range keyboardButtons {
						list += fmt.Sprintf("%d. %s\n", i+1, btn)
					}
					sendMessage(userID, "🗑️ شماره دکمه‌ای که می‌خوای حذف کنی رو وارد کن:\n\n"+list, false)
					continue

				case "📊 آمار کاربران":
					sendMessage(userID, "📊 آمار کاربران ahb panel:\n\n👥 تعداد کل کاربران: ۰\n🟢 آنلاین: ۰", false)
					continue

				case "🔙 بازگشت":
					sendMessage(userID, "🔙 به منوی اصلی برگشتی.", false)
					continue
				}

				if strings.HasPrefix(text, "💳") && text != "💳 تغییر شماره کارت" {
					cardNumber = text
					sendMessage(userID, "✅ شماره کارت با موفقیت تغییر کرد:\n\n`"+cardNumber+"`", true)
					continue
				}

				if strings.HasPrefix(text, "➕") && text != "➕ افزودن دکمه" {
					newButton := strings.TrimPrefix(text, "➕ ")
					if newButton != "" {
						keyboardButtons = append(keyboardButtons, newButton)
						sendMessage(userID, "✅ دکمه `"+newButton+"` با موفقیت اضافه شد.", true)
					}
					continue
				}

				if idx, err := strconv.Atoi(text); err == nil && idx > 0 && idx <= len(keyboardButtons) {
					removed := keyboardButtons[idx-1]
					keyboardButtons = append(keyboardButtons[:idx-1], keyboardButtons[idx:]...)
					sendMessage(userID, "✅ دکمه `"+removed+"` با موفقیت حذف شد.", true)
					continue
				}
			}
		}
	}
}

func sendMessage(chatID int64, text string, isMarkdown bool) {
	msg := tgbotapi.NewMessage(chatID, text)
	if isMarkdown {
		msg.ParseMode = tgbotapi.ModeMarkdown
	}
	if !isMarkdown {
		msg.ReplyMarkup = getKeyboard()
	}
	bot.Send(msg)
}

func getKeyboard() tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for _, btn := range keyboardButtons {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btn)))
	}
	if ownerID != 0 {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⚙️ تنظیمات"),
			tgbotapi.NewKeyboardButton("📊 آمار کاربران"),
		))
	}
	return tgbotapi.NewReplyKeyboard(rows...)
}

func sendAdminPanel(chatID int64) {
	text := "⚙️ **پنل مدیریت ahb panel**\n\n"
	text += "🔹 **شماره کارت فعلی:** " + cardNumber + "\n"
	text += "🔹 **تعداد دکمه‌ها:** " + strconv.Itoa(len(keyboardButtons)) + "\n\n"
	text += "از دکمه‌های زیر برای مدیریت استفاده کن:"

	adminKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ افزودن دکمه"),
			tgbotapi.NewKeyboardButton("🗑️ حذف دکمه"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💳 تغییر شماره کارت"),
			tgbotapi.NewKeyboardButton("📊 آمار کاربران"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 بازگشت"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = adminKeyboard
	bot.Send(msg)
}
