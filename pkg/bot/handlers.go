package bot

import (
	"context"
	"fmt"
	"gradebot/pkg/db"
	"gradebot/pkg/embedlog"
	"math"
	"math/rand"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	scholarPercent  = 100
	brokePercent    = 250
	internPercent   = 350
	juniorPercent   = 650
	middlePercent   = 850
	seniorPercent   = 920
	teamLeadPercent = 970
	ceoPercent      = 985
	papikPercent    = 995
	mayatinPercent  = 1000
)

var salariesMap = map[int]string{
	scholarPercent:  `Ты школота\, копишь на обеды\, прогаешь только домашнее задание 🫵😹 Зарплату получаешь от мамы\, премию \- от бабушки\.`,
	brokePercent:    `Ты безработный 🫵😹 Стипендии и денег родителей пока что хватает на еду\, но я тебе не завидую :/`,
	internPercent:   `Тебя взяли стажёром в твою первую IT\-галеру 🔥 Теперь ты \- настоящий программист\! Правда\, придётся ишачить 2 года\, чтобы получить повышение\.\.\.`,
	juniorPercent:   `Ты получаешь гордое звание джуна\! 🧑‍💻 Таких как ты \- абсолютное большинство\. Удачи пробиться на Хедхантере :\)`,
	middlePercent:   `Ты \- мидл\! 🤠 Молодец\, немногие сюда добираются\. А теперь настало время выгорания на работе 🔥🔥🔥`,
	seniorPercent:   `Ты \- сеньор\! 🤑 Можешь работать по 3 часа в день\, хрюшам все равно дороже искать замену`,
	teamLeadPercent: `Ты \- тимлид\! 👨‍💼 Можешь вообще не работать\, а сидеть на созвонах и важных встречах целый день`,
	ceoPercent:      `Ты \- CEO\! 😎 Пока эти лошпеды тратят нервы на кодинг\, ты получаешь все сливки с их трудов\. Все потому что ты \- лучше\, чем они\. Не забывай напоминать им об этом\!`,
	papikPercent:    `Ты \- Папикян Сергей Седракович\, легенда ИТМО и самый богатый человек в мире\. Ты победил в этой жизни\, все тебе завидуют\.`,
	mayatinPercent:  `Ты \- Маятин Александр Владимирович\! Три источника твоего богатства \- Производительность\, Надежность и Безопасность\.`,
}

type UndefinedSong struct {
	Title  string
	FileID string
}

var undefinedSongs = []UndefinedSong{
	{Title: "👨‍🦳 Я Папикян\nТы настоящий олд!", FileID: "CQACAgIAAxkBAAMSZ_ziMUWAFmAoHVERRnASZECYR1EAAoMqAAKKY3lLl1qro2901rg2BA"},
	{Title: "🇧🇷🇧🇷🇧🇷 BRAZZIL\nТы трушный студент ФИТИП, возможно пора задуматься о путешествии по Южной Америке?🤔", FileID: "CQACAgIAAxkBAAMeZ_ziMdVaysbZ9povh2-VSAABqND8AAJVZQACTYCBSAGrlmTsAAH9TDYE"},
	{Title: "🤠 morgenISIT\nТы познал все тяжести ИС и теперь целыми днями переписываешься с такими же старичками в оффтопе, вспоминая былое", FileID: "CQACAgIAAxkBAAMTZ_ziMd0xgFjMfIZNVUoxgAzlxroAAhUmAAKKY4FLyQJtRjp6jck2BA"},
	{Title: "😎 BlackPapik [Папин Танк]\nТы уважаешь Сергея Седраковича больше, чем все остальные!", FileID: "CQACAgIAAxkBAAMUZ_ziMex1TFiiAAHn9x3VQKNqEcJSAAIhJgACimOBS_ujBMUBINiXNgQ"},
	{Title: "🗿 ballad\nТы говоришь на языке фактов, продолжай в том же духе", FileID: "CQACAgIAAxkBAAMZZ_ziMZOLGPy9Ha_4MlvjenmztXcAArc1AAIB5-FL_XuDq8CnmlA2BA"},
	{Title: "🖐✌️ +7(952)09-03-02\nНет слов, только 52", FileID: "CQACAgIAAxkBAAMaZ_ziMRPjf-MPtN3ZExEvOAnueAMAArZDAALJz-lJ-9xef8n3bTE2BA"},
	{Title: "👨‍💻 OOP [Nominalo]\nТы проводишь все выходные без сна, переписывая лабы после очередных правок. Зато потом будешь экспертом по ООП!", FileID: "CQACAgIAAxkBAAMVZ_ziMetBWvbzfiLSyCu-6RNlSG0AAiYmAAKKY4FL-eztebHcyeo2BA"},
	{Title: "😔 Kreed\nТы сегодня в меланхолично-депрессивном вайбе. Не подходи к балконам и открытым окнам", FileID: "CQACAgIAAxkBAAMWZ_ziMRAox71pFkvA-1RI29O180sAAhonAAKKY4FLGx8iCoZbEmY2BA"},
	{Title: "😍🥰😮‍💨 NE ROMA\nТы самый трушный фан Ромы!!!", FileID: "CQACAgIAAxkBAAMXZ_ziMR_Z64tpnb1UX-S4jlAHC6IAAgkrAAKMq3hJRnGZyQABhBjcNgQ"},
	{Title: "☺️😌😘 heronwater\nНаступила весна и у тебя настроение влюбляться!", FileID: "CQACAgIAAxkBAAMYZ_ziMYDbUgYdZEIK7IIoJbAYuxMAAlM2AALiFghL4VMu9ITVwxs2BA"},
	{Title: "❤️‍🩹❤️‍🩹❤️‍🩹 fitp juice wrld\nВремя выйти на балкон, закурить, задуматься обо всём, что было за эти годы...", FileID: "CQACAgIAAxkBAAMbZ_ziMWNoqFshqo_s2KLo8JUowPIAAvNOAALBJiBLUzTze2mBAAHENgQ"},
	{Title: "🕺💃 Ronimizy\nПолторы минуты вайба в перерыве между сотней лаб - вот всё, что тебе светит в этом семестре", FileID: "CQACAgIAAxkBAAMcZ_ziMZmNwctLX2pXSnfnT69X5c8AAq9XAAJ15nBKNGA4zqckH5w2BA"},
	{Title: "🤖Бонусный нейротрек!\nКажется, тебе пора задуматься о бэкапах. Посмотри внимательно, не упала ли ещё база на проде?", FileID: "CQACAgIAAxkBAAMoZ_5J-6BkTYKCUtAjK_Y8-gWtsXAAAhBqAAKEnPlLivuOrk1ND6I2BA"},
	{Title: "🤖Бонусный нейротрек!\nВ ближайшее время тебе светит освобождение от занятий! Правда, за ним последует очень сложный экзамен...", FileID: "CQACAgIAAxkBAAMjZ_5J-6UDE84tD3fV0Met9PaO-80AAgtqAAKEnPlLP7MOBeTmt7I2BA"},
	{Title: "🤖Бонусный нейротрек!\nМаксим Валерьевич сегодня не в духе, не советую приближаться к кустам!", FileID: "CQACAgIAAxkBAAMkZ_5J-09afhCqAAHpKEVNife46VjIAAIMagAChJz5S7J1rUPi_KF4NgQ"},
	{Title: "🤖Бонусный нейротрек!\nСамое время садиться писать диплом, сроки уже горят!", FileID: "CQACAgIAAxkBAAMlZ_5J-yaHqnazWWOdEAABAqH27iJbAAINagAChJz5SxbigaUC0GNXNgQ"},
	{Title: "🤖Бонусный нейротрек!\nВся правда о взаимоотношениях этой троицы - только в этом аудиофайле", FileID: "CQACAgIAAxkBAAMmZ_5J-8KxhV_FYAQQMFtKqV4FKGYAAg5qAAKEnPlLZxI20ZJi1Ok2BA"},
	{Title: "🤖Бонусный нейротрек!\nСегодня в тебе много агрессии и раздражения, пора расслабиться, в кальянную сходи хз", FileID: "CQACAgIAAxkBAAMnZ_5J-2Nvq9xoMFDKwd0f5btGbVEAAg9qAAKEnPlLyLq_WQEys4E2BA"},
	{Title: "🤖Бонусный нейротрек!\nУ тебя сегодня крупные косяки... Придётся отрабатывать", FileID: "CQACAgIAAxkBAAMiZ_5J-0IG8ypyyPEDeNzSiEfNtKIAAglqAAKEnPlLhvzbrfIpA702BA"},
	{Title: "🍾🔥😮‍💨 Я закрыл UML (секретный дроп!)\nВздохни с облегчением, худшее уже позади :)", FileID: "CQACAgIAAxkBAAMsZ_5p16mtBtMNY8oOYzpOh7LRHsIAAnZLAAJfOClJ0ll_antjt_k2BA"},
}

type BotService struct {
	embedlog.Logger
	db db.DB
}

func NewBotService(logger embedlog.Logger, db db.DB) *BotService {
	return &BotService{Logger: logger, db: db}
}

func (bs BotService) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil && update.Message.Audio != nil {
		println(update.Message.Audio.FileName, "|", update.Message.Audio.FileID)
	}
	if update.InlineQuery != nil && update.InlineQuery.From != nil {
		if err := bs.answerInlineQuery(ctx, b, update); err != nil {
			bs.Errorf("%v", err)
		}
		return
	}
	return
}

func (bs BotService) answerInlineQuery(ctx context.Context, b *bot.Bot, update *models.Update) error {

	var salary int
	var ending string
	percents := rand.Intn(1000)
	switch {
	case percents <= scholarPercent:
		salary = rand.Intn(1000)
		ending = salariesMap[scholarPercent]
		break
	case percents <= brokePercent:
		salary = 10000 + rand.Intn(20000)
		ending = salariesMap[brokePercent]
		break
	case percents <= internPercent:
		salary = 20000 + (rand.Intn(25000)/1000)*1000
		ending = salariesMap[internPercent]
		break
	case percents <= juniorPercent:
		salary = 40000 + (rand.Intn(40000)/1000)*1000
		ending = salariesMap[juniorPercent]
		break
	case percents <= middlePercent:
		salary = 80000 + (rand.Intn(200000)/5000)*5000
		ending = salariesMap[middlePercent]
		break
	case percents <= seniorPercent:
		salary = 280000 + (rand.Intn(320000)/10000)*10000
		ending = salariesMap[seniorPercent]
		break
	case percents <= teamLeadPercent:
		salary = 600000 + (rand.Intn(500000)/50000)*50000
		ending = salariesMap[teamLeadPercent]
		break
	case percents <= ceoPercent:
		salary = 1100000 + (rand.Intn(10000000)/100000)*100000
		ending = salariesMap[ceoPercent]
		break
	case percents <= papikPercent:
		ending = salariesMap[papikPercent]
		salary = math.MaxInt32
		break
	case percents <= mayatinPercent:
		ending = salariesMap[mayatinPercent]
		salary = math.MaxInt32
		break
	}

	username := update.InlineQuery.From.Username
	username = strings.ReplaceAll(username, "_", `\_`)
	username = strings.ReplaceAll(username, "!", `\!`)
	username = strings.ReplaceAll(username, ".", `\.`)
	username = strings.ReplaceAll(username, ",", `\,`)
	username = strings.ReplaceAll(username, `-`, `\-`)
	username = strings.ReplaceAll(username, `=`, `\=`)
	username = strings.ReplaceAll(username, `#`, `\#`)
	username = strings.ReplaceAll(username, `+`, `\+`)
	username = strings.ReplaceAll(username, `(`, `\(`)
	username = strings.ReplaceAll(username, `)`, `\)`)
	username = strings.ReplaceAll(username, `*`, `\*`)
	username = strings.ReplaceAll(username, `~`, `\~`)
	username = strings.ReplaceAll(username, `[`, `\[`)
	username = strings.ReplaceAll(username, `]`, `\]`)

	// send answer to the query
	results := []models.InlineQueryResult{
		&models.InlineQueryResultArticle{
			ID:           "1",
			Title:        "Твоя зп",
			ThumbnailURL: "https://cdn.vectorstock.com/i/500p/79/20/emoticon-with-dollars-vector-2287920.jpg",
			ReplyMarkup: models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{
						models.InlineKeyboardButton{
							Text:                         "Узнать свою",
							SwitchInlineQueryCurrentChat: " ",
						},
					},
				}},
			InputMessageContent: &models.InputTextMessageContent{
				MessageText: fmt.Sprintf("Зарплата @%s: ||%d₽\n%s||", username, salary, ending),
				ParseMode:   models.ParseModeMarkdown,
			}},
		&models.InlineQueryResultArticle{
			ID:           "2",
			Title:        "Кто ты из песен Undefined?",
			ThumbnailURL: "https://memi.klev.club/uploads/posts/2024-12/memi-klev-club-ngvi-p-memi-s-devushkoi-v-naushnikakh-1.jpg",
			ReplyMarkup: models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{
						models.InlineKeyboardButton{
							Text:         "Узнать песню",
							CallbackData: "song",
						},
					},
				}},
			InputMessageContent: &models.InputTextMessageContent{
				MessageText: "Нажми и узнаешь!",
			},
		},
	}

	_, err := b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		//Button: &models.InlineQueryResultsButton{
		//	Text:           "Оставить фидбек",
		//	StartParameter: "1",
		//},
		InlineQueryID: update.InlineQuery.ID,
		Results:       results,
		IsPersonal:    true,
		CacheTime:     1,
	})

	return err
}

func FindUndefinedSong(ctx context.Context, b *bot.Bot, update *models.Update) {
	selectedSong := undefinedSongs[rand.Intn(len(undefinedSongs))]
	b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaAudio{
			Media:   selectedSong.FileID,
			Caption: selectedSong.Title,
		},
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					models.InlineKeyboardButton{
						Text:                         "Узнать свою",
						SwitchInlineQueryCurrentChat: " ",
					},
				},
			}},
	})
}
