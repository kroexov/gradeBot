package bot

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-pg/pg/v10"
	"gradebot/pkg/db"
	"gradebot/pkg/embedlog"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	patternPapikSlots         = "papikSlots"
	patternMayatinRoulette    = "mayatinRoulette"
	patternMayatinRouletteBet = "mayatinBet"
	patternPovyshevExams      = "povyshevExams"
	patternBuyBack            = "buyback"
	playersRating             = "rating"

	patternMayatinRouletteBetN = "_n"
	patternMayatinRouletteBetP = "_p"
	patternMayatinRouletteBetB = "_b"
	patternMayatinRouletteBetU = "_u"
)

var slotsResults = [7]string{
	"https://i.ibb.co/1YqJpXwW/photo-2025-03-21-18-45-11.jpg",
	"https://i.ibb.co/jPJ6TJ7Q/photo-2025-03-21-18-45-14.jpg",
	"https://i.ibb.co/Z6PhZ8jh/photo-2025-03-21-18-45-17.jpg",
	"https://i.ibb.co/qYLRLcN0/photo-2025-03-21-18-45-19.jpg",
	"https://i.ibb.co/m5Ykp15w/photo-2025-03-21-18-45-22.jpg",
	"https://i.ibb.co/pBYcBbDJ/photo-2025-03-21-18-45-25.jpg",
	"https://i.ibb.co/rRBVsQJC/photo-2025-03-21-18-45-27.jpg",
}

type MayatinRouletteCategory struct {
	CategoryName string
	CategoryPic  string
	WinSum       int
}

var mayatinCategories = map[string]MayatinRouletteCategory{
	patternMayatinRouletteBetN: {
		CategoryName: "Надежность",
		CategoryPic:  "https://i.ibb.co/mCxMpSdk/image.png",
		WinSum:       300000,
	},
	patternMayatinRouletteBetP: {
		CategoryName: "Производительность",
		CategoryPic:  "https://i.ibb.co/Zpqh23VB/image.png",
		WinSum:       300000,
	},
	patternMayatinRouletteBetB: {
		CategoryName: "Безопасность",
		CategoryPic:  "https://i.ibb.co/WNbKBsrp/image.png",
		WinSum:       300000,
	},
	patternMayatinRouletteBetU: {
		CategoryName: "Уважаемый коллега",
		CategoryPic:  "https://i.ibb.co/DPjH6ym5/image.png",
		WinSum:       1000000,
	},
}

type BotService struct {
	embedlog.Logger
	db db.DB

	cr                      db.CommonRepo
	mayatinRouletteBets     *sync.Map
	isMayatinRouletteActive bool
	mayatinRouletteUsers    map[int]struct{}
}

func NewBotService(logger embedlog.Logger, dbo db.DB) *BotService {
	return &BotService{Logger: logger, db: dbo, cr: db.NewCommonRepo(dbo), mayatinRouletteBets: new(sync.Map)}
}

func (bs *BotService) RegisterBotHandlers(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternPapikSlots, bot.MatchTypePrefix, bs.PapikRouletteHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternMayatinRoulette, bot.MatchTypePrefix, bs.MayatinRouletteHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternMayatinRouletteBet, bot.MatchTypePrefix, bs.MayatinRouletteBetHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, playersRating, bot.MatchTypePrefix, bs.PlayersRatingHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, patternBuyBack, bot.MatchTypePrefix, bs.BuyBackHandler)
}

func (bs BotService) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.InlineQuery != nil && update.InlineQuery.From != nil {
		if err := bs.answerInlineQuery(ctx, b, update); err != nil {
			bs.Errorf("%v", err)
		}
		return
	}
	return
}

func (bs BotService) answerInlineQuery(ctx context.Context, b *bot.Bot, update *models.Update) error {
	username := update.InlineQuery.From.Username
	user, err := bs.cr.OneLudoman(ctx, &db.LudomanSearch{LudomanNickname: &username})
	if err != nil {
		return err
	}
	if user == nil {
		newUser, err := bs.cr.AddLudoman(ctx, &db.Ludoman{
			LudomanNickname: username,
			Balance:         1000000,
		})
		if err != nil {
			return err
		}
		// send answer to the query
		_, err = b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
			//Button: &models.InlineQueryResultsButton{
			//	Text:           "Оставить фидбек",
			//	StartParameter: "1",
			//},
			InlineQueryID: update.InlineQuery.ID,
			Results: []models.InlineQueryResult{
				&models.InlineQueryResultArticle{
					ID:           "1",
					Title:        "Добро пожаловать!",
					ThumbnailURL: "https://i.ibb.co/Xfx3C5wH/image-1.jpg",
					ReplyMarkup: models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								models.InlineKeyboardButton{
									Text:         "Слоты Папикяна",
									CallbackData: patternPapikSlots + "_" + strconv.Itoa(newUser.ID),
								},
							},
							{
								models.InlineKeyboardButton{
									Text:         "Рулетка Маятина",
									CallbackData: patternMayatinRoulette + "_" + strconv.Itoa(newUser.ID),
								},
							},
							{
								models.InlineKeyboardButton{
									Text:         "Экзамен Повышева",
									CallbackData: patternPovyshevExams + "_" + strconv.Itoa(newUser.ID),
								},
							},
						}},
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: fmt.Sprintf("Добро пожаловать в И$ - Казик, @%s!\nВам начислен 1.000.000 I$Coins за первый визит. Выбирайте игру и побеждайте!", username),
					}},
			},
			IsPersonal: true,
			CacheTime:  1,
		})
	} else {
		_, err = b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
			//Button: &models.InlineQueryResultsButton{
			//	Text:           "Оставить фидбек",
			//	StartParameter: "1",
			//},
			InlineQueryID: update.InlineQuery.ID,
			Results: []models.InlineQueryResult{
				&models.InlineQueryResultArticle{
					ID:           "1",
					Title:        "Выберите игру!",
					ThumbnailURL: "https://i.ibb.co/Xfx3C5wH/image-1.jpg",
					ReplyMarkup: models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								models.InlineKeyboardButton{
									Text:         "Слоты Папикяна",
									CallbackData: patternPapikSlots + "_" + strconv.Itoa(user.ID),
								},
							},
							{
								models.InlineKeyboardButton{
									Text:         "Рулетка Маятина",
									CallbackData: patternMayatinRoulette + "_" + strconv.Itoa(user.ID),
								},
							},
							{
								models.InlineKeyboardButton{
									Text:         "Экзамен Повышева (в разработке)",
									CallbackData: patternPovyshevExams + "_" + strconv.Itoa(user.ID),
								},
							},
						}},
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: fmt.Sprintf("Добро пожаловать в И$ - Казик, @%s!\nВыбирайте игру и побеждайте!", username),
					}},
				&models.InlineQueryResultArticle{
					ID:           "2",
					Title:        "Рейтинг игроков!",
					ThumbnailURL: "https://russia-rating.ru/wp-content/uploads/2024/09/567.jpg",
					ReplyMarkup: models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								models.InlineKeyboardButton{
									Text:         "Узнать рейтинг",
									CallbackData: playersRating,
								},
							},
						}},
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: fmt.Sprintf("Добро пожаловать в И$ - Казик, @%s!\nНажмите кнопку ниже, чтобы узнать рейтинг игроков!", username),
					}},
				&models.InlineQueryResultArticle{
					ID:           "3",
					Title:        "Правила",
					ThumbnailURL: "https://casino.ru/wp-content/uploads/articles/poker/poker-1-400x266.jpg",
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: fmt.Sprintf("Добро пожаловать в И$ - Казик, @%s!\nВот список наших развлечений:\n1. Слоты Папикяна. Вход 100.000, шанс на выигрыш 1/7, размер выигрыша 500.000\n2. Рулетка Маятина. Вход 100.000, шансы на выигрыш: 3/10 с возвратом 300.000, либо 1/10 с возвратом 1.000.000\n3. Экзамен Повышева (в разработке). Вход 100.000, шансы на выигрыш 1/6 в размере 500.000, либо взять седьмой \"удачный билет\" с шансом 50/50 и выигрышем 500.000, но ставкой 300.000\n\nВо всех автоматах есть 1/100 шанс на Гигавыигрыш в размере 10.000.000! (в разработке)", username),
					}},
			},
			IsPersonal: true,
			CacheTime:  1,
		})
	}

	return err
}

func (bs *BotService) PapikRouletteHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, "_")
	if len(parts) < 2 {
		bs.Errorf("len(parts) < 2")
		return
	}

	// find user
	userId, err := strconv.Atoi(parts[1])
	if err != nil {
		bs.Errorf("%v", err)
	}
	user, err := bs.cr.LudomanByID(ctx, userId)
	if err != nil {
		bs.Errorf("%v", err)
	}

	if user.LudomanNickname != update.CallbackQuery.From.Username {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Это не ваш автомат! Нажмите на название бота и тоже сможете сыграть :)",
			ShowAlert:       true,
		})
		return
	}

	if user.Balance < 100000 {
		bs.lossHandler(ctx, b, update, parts[1])
		return
	}

	b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaAnimation{
			Media:     "https://media.tenor.com/_yoDqyYP8aYAAAAM/casino77-slot-machine.gif",
			Caption:   "Крутимся...",
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
	})

	time.Sleep(3 * time.Second)

	num := rand.Intn(len(slotsResults))
	var res string
	switch num {
	case 0:
		user.Balance += 500000
		res = fmt.Sprintf("@%s, Победа! Вы получаете +500.000 I$Coins. Ваш текущий баланс: %d I$Coins", update.CallbackQuery.From.Username, user.Balance)
	default:
		user.Balance -= 100000
		res = fmt.Sprintf("@%s, Неудача! Ваш текущий баланс: %d I$Coins", update.CallbackQuery.From.Username, user.Balance)
	}

	pic := slotsResults[num]

	b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaPhoto{
			Media:     pic,
			Caption:   res,
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				models.InlineKeyboardButton{
					Text:         "Сыграть ещё раз",
					CallbackData: patternPapikSlots + "_" + parts[1],
				},
			},
		}},
	})

	_, err = bs.cr.UpdateLudoman(ctx, user, db.WithColumns(db.Columns.Ludoman.Balance))
	if err != nil {
		bs.Errorf("%v", err)
		return
	}
}

func (bs *BotService) lossHandler(ctx context.Context, b *bot.Bot, update *models.Update, userId string) {
	b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaPhoto{
			Media:     "https://i.ibb.co/8C2G9X9/image.png",
			Caption:   "Вы израходовали свой баланс!",
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				models.InlineKeyboardButton{
					Text:         "Хочу откупиться!",
					CallbackData: patternBuyBack + "_" + userId,
				},
			},
		}},
	})
}

func (bs *BotService) PlayersRatingHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	players, err := bs.cr.LudomenByFilters(ctx, &db.LudomanSearch{}, db.Pager{PageSize: 10}, db.WithSort(db.NewSortField(db.Columns.Ludoman.Balance, true)))
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	// Шаблон для вывода списка
	listTemplate := `Список игроков:
{{- range $index, $ludoman := . }}
{{- printf "\n%d. Никнейм: @%s, Баланс: %d, Всего проигрышей: %d" (add $index 1) $ludoman.LudomanNickname $ludoman.Balance $ludoman.Losses}}
{{- end }}
`
	// Функция для добавления 1 к индексу (так как индексация с 0)
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	// Создаем шаблон и парсим его
	tmpl, err := template.New("list").Funcs(funcMap).Parse(listTemplate)
	if err != nil {
		bs.Errorf("%v", err)
	}

	var buf bytes.Buffer

	// Выполняем шаблон и выводим результат
	err = tmpl.Execute(&buf, players)
	if err != nil {
		bs.Errorf("%v", err)
	}

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Text:            buf.String(),
	})
	if err != nil {
		bs.Errorf("%v", err)
	}
}

func (bs *BotService) BuyBackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, "_")
	if len(parts) < 2 {
		bs.Errorf("len(parts) < 2")
		return
	}

	// find user
	userId, err := strconv.Atoi(parts[1])
	if err != nil {
		bs.Errorf("%v", err)
	}
	user, err := bs.cr.LudomanByID(ctx, userId)
	if err != nil {
		bs.Errorf("%v", err)
	}

	if user.LudomanNickname != update.CallbackQuery.From.Username {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Это не ваш автомат! Нажмите на название бота и тоже сможете сыграть :)",
			ShowAlert:       true,
		})
		if err != nil {
			bs.Errorf("%v", err)
		}
		return
	}

	user.Balance = 1000000
	user.Losses += 1
	_, err = bs.cr.UpdateLudoman(ctx, user, db.WithColumns(db.Columns.Ludoman.Balance, db.Columns.Ludoman.Losses))
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaPhoto{
			Media:     "https://i.ibb.co/6R0Cz78Q/image-4.jpg",
			Caption:   fmt.Sprintf("Вы откупились! Счетчик ваших проданных квартир: %d\nНажмите на название бота и проиграйте всё снова, или может быть сегодня вам повезет попасть в топ рейтинга?)", user.Losses),
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
	})
}

func (bs *BotService) MayatinRouletteHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, "_")
	if len(parts) < 2 {
		bs.Errorf("len(parts) < 2")
		return
	}

	// find user
	userId, err := strconv.Atoi(parts[1])
	if err != nil {
		bs.Errorf("%v", err)
	}
	user, err := bs.cr.LudomanByID(ctx, userId)
	if err != nil {
		bs.Errorf("%v", err)
	}

	if user.LudomanNickname != update.CallbackQuery.From.Username {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Это не ваш автомат! Нажмите на название бота и тоже сможете сыграть :)",
			ShowAlert:       true,
		})
		return
	}

	if user.Balance < 100000 {
		bs.lossHandler(ctx, b, update, parts[1])
		return
	}

	if bs.isMayatinRouletteActive {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Рулетка уже идет! Присоединяйтесь к текущей!",
			ShowAlert:       true,
		})
		return
	}

	bs.mayatinRouletteBets = new(sync.Map)
	bs.isMayatinRouletteActive = true
	bs.mayatinRouletteUsers = make(map[int]struct{})
	bs.mayatinRouletteBets.Store(patternMayatinRoulette, map[string][]int{})

	for i := 0; i < 15; i++ {
		v, ok := bs.mayatinRouletteBets.Load(patternMayatinRoulette)
		if !ok {
			bs.Errorf("not found syncMap")
			return
		}

		bets, ok := v.(map[string][]int)
		if !ok {
			bs.Errorf("can't convert bets")
			return
		}

		b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
			InlineMessageID: update.CallbackQuery.InlineMessageID,
			Media: &models.InputMediaAnimation{
				Media:     "https://i.pinimg.com/originals/32/37/bf/3237bf1e172a6089e0c437ffd3b28010.gif",
				Caption:   fmt.Sprintf("Рулетка Маятина началась! Выбирайте ваш слот в рулетке!\nСтавка 100.000, слот 'Уважаемый коллега дает 10x выигрыш, но выпадает реже'\nОсталось %d секунд!", 15-i),
				ParseMode: models.ParseModeHTML,
				//HasSpoiler: true,
			},
			ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					models.InlineKeyboardButton{
						Text:         fmt.Sprintf("Надёжность! (%d ставок)", len(bets[patternMayatinRouletteBetN])),
						CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetN,
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         fmt.Sprintf("Производительность! (%d ставок)", len(bets[patternMayatinRouletteBetP])),
						CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetP,
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         fmt.Sprintf("Безопасность! (%d ставок)", len(bets[patternMayatinRouletteBetB])),
						CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetB,
					},
				},
				{
					models.InlineKeyboardButton{
						Text:         fmt.Sprintf("Уважаемый коллега 😎 (10x выигрыш, %d ставок)", len(bets[patternMayatinRouletteBetU])),
						CallbackData: patternMayatinRouletteBet + patternMayatinRouletteBetU,
					},
				},
			}},
		})
		time.Sleep(1 * time.Second)
	}

	i := rand.Intn(100)
	var selectedCategory string
	switch {
	case i <= 30:
		selectedCategory = patternMayatinRouletteBetP
	case i <= 60:
		selectedCategory = patternMayatinRouletteBetB
	case i <= 90:
		selectedCategory = patternMayatinRouletteBetN
	default:
		selectedCategory = patternMayatinRouletteBetU
	}
	cat := mayatinCategories[selectedCategory]

	v, ok := bs.mayatinRouletteBets.Load(patternMayatinRoulette)
	if !ok {
		bs.Errorf("not found syncMap")
		return
	}

	bets, ok := v.(map[string][]int)
	if !ok {
		bs.Errorf("can't convert bets")
		return
	}

	var result string
	winners := bets[selectedCategory]
	if len(winners) == 0 {
		result = `Победителей нет 🫵😹`
	} else {
		winUsers, err := bs.cr.LudomenByFilters(ctx, &db.LudomanSearch{IDs: winners}, db.PagerNoLimit)
		if err != nil {
			bs.Errorf("%v", err)
		}
		result = "Список победителей: "
		for _, winUser := range winUsers {
			result += "@" + winUser.LudomanNickname + " "
		}
		result += fmt.Sprintf("\nПобедителям начислено: %d", cat.WinSum)

		err = bs.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			for _, winUser := range winUsers {
				winUser.Balance += cat.WinSum
				_, err = bs.cr.UpdateLudoman(ctx, &winUser, db.WithColumns(db.Columns.Ludoman.Balance))
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			bs.Errorf("%v", err)
			return
		}
	}

	b.EditMessageMedia(ctx, &bot.EditMessageMediaParams{
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		Media: &models.InputMediaPhoto{
			Media:     cat.CategoryPic,
			Caption:   fmt.Sprintf("Рулетка Маятина завершена! Выпало: %s!\n%s", cat.CategoryName, result),
			ParseMode: models.ParseModeHTML,
			//HasSpoiler: true,
		},
	})

	bs.isMayatinRouletteActive = false
}

func (bs *BotService) MayatinRouletteBetHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, "_")
	if len(parts) < 2 {
		bs.Errorf("len(parts) < 2")
		return
	}
	userBet := parts[1]

	// find user
	user, err := bs.cr.OneLudoman(ctx, &db.LudomanSearch{LudomanNickname: &update.CallbackQuery.From.Username})
	if err != nil {
		bs.Errorf("%v", err)
	} else if user == nil {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Вас еще нет в нашей базе данных :( Попробуйте сначала зарегаться в боте",
			ShowAlert:       true,
		})
		return
	}

	if _, ok := bs.mayatinRouletteUsers[user.ID]; ok {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Вы уже поставили на рулетку! Теперь ждите и молитесь :)",
			ShowAlert:       true,
		})
		return
	}

	if user.Balance < 100000 {
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "У вас недостаточно денег для этой ставки :/",
			ShowAlert:       true,
		})
		return
	}

	if bs.mayatinRouletteUsers == nil {
		return
	}

	bs.mayatinRouletteUsers[user.ID] = struct{}{}
	v, ok := bs.mayatinRouletteBets.Load(patternMayatinRoulette)
	if !ok {
		bs.Errorf("not found syncMap")
		return
	}

	bets, ok := v.(map[string][]int)
	if !ok {
		bs.Errorf("can't convert bets")
		return
	}

	user.Balance -= 100000
	_, err = bs.cr.UpdateLudoman(ctx, user)
	if err != nil {
		bs.Errorf("%v", err)
		return
	}

	newBets := append(bets["_"+userBet], user.ID)
	bets["_"+userBet] = newBets
	bs.mayatinRouletteBets.Store(patternMayatinRoulette, bets)
}

func Pointer[T any](in T) *T {
	return &in
}
