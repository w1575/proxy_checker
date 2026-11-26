package main

const (
	langKey = "lang"
	LangRU  = "ru"
	LangEn  = "en"
)

func transFn() func(key string) string {
	return func(key string) string {
		str, ok := en[key]
		if !ok || settings[langKey] == LangRU {
			return key
		}

		return str
	}
}

var en = map[string]string{
	"Что хотите сделать?": "What would you like to do?",
	"0. Выход":            "0. Exit",
	"2. Проверить raw.txt (все прокси должны быть руками записаны)":       "2. Check raw.txt (all proxies must be entered manually)",
	"1. Скачать все прокси из файла источников (sourcess.txt -> raw.txt)": "1. Download all proxies from the source file (sources.txt -> raw.txt)",
	"3. Скачать все прокси из файла источников и автоматически проверить": "3. Download all proxies from the source file and check them automatically",

	"Скачиваем файл с гитхаба.":          "Download the file from GitHub.",
	"Начинаем проверку прокси серверов.": "Begin checking the proxy servers.",

	"Начали скачивать файл со списком прокси с гитхаба.": "Started downloading the file containing the list of proxies from GitHub.",
	"при скачивании списка прокси произошла ошибка: %s":  "An error occurred while downloading the proxy list: %s",
	"При записи прокси в файл произошла ошибка: %s":      "An error occurred while writing the proxies to a file: %s",

	"увы, получен скорее всего пустой ответ": "unfortunately, the response received is most likely empty",
	"Получили ответ от API %s":               "Received a response from API %s",

	"Raw file не удалось открыть":            "Failed to open the raw file",
	"Прошло времени с момента запуска %d \n": "Time elapsed since launch: %d \n",
}
