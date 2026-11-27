# *Proxy Checker (ALPHA VERSION)*

[🇬🇧 English](#what-is-it)

## *Что это такое?*

Небольшая утилита, которая проверяет список прокси из файла `raw.txt` и записывает рабочие прокси в `valid.txt`, а нерабочие — в `invalid.txt`.

Этот небольшой скрипт изначально был частью другого моего проекта (оказавшегося слишком амбициозным для меня). Чтобы не дать ему пропасть, я решил выделить его в отдельную небольшую утилиту.

Также имеется экспериментальная функция: программа самостоятельно скачивает и проверяет списки прокси из репозитория [iplocate/free-proxy-list](https://github.com/iplocate/free-proxy-list).

Небольшой нюанс: многие прокси из этого репозитория, несмотря на то что технически работают, заблокированы на множестве сайтов. Как говорится, бесплатный сыр бывает только в мышеловке.

Пока что единственная функция файла `settings.yml` — переключение языка.

## *В планах:*

- ~~Добавить возможность загрузки адресов из разных источников.~~
- ~~Добавить Makefile для сборки проекта в отдельную директорию.~~
- Добавить возможность парсить исходные файлы в разных форматах (изначально программа была рассчитана исключительно на [all-proxies.txt](https://raw.githubusercontent.com/iplocate/free-proxy-list/refs/heads/main/all-proxies.txt)). В идеале пользователь сможет самостоятельно указывать формат каждого источника. Или, возможно, сделать автоматическое определение?
- Добавить дополнительные настройки количества потоков в `settings.yml`.

## *Сборка*

Единственное требование — установленный Go версии 1.27.x или новее.

Для быстрой сборки достаточно выполнить команду `make` в терминале. Если утилита Make не установлена, можно воспользоваться соответствующей командой для своей операционной системы из файла `Makefile`.

После сборки в корне проекта будет создана директория `build`, содержащая следующие файлы:

- `raw.txt`
- `valid.txt`
- `invalid.txt`
- `sources.txt`
- `settings.yml`
- Исполняемый файл `proxy_checker` (или `proxy_checker.exe`).

**Важно:** сборка под Windows пока не тестировалась, поэтому её работоспособность не гарантируется.

---

## *What is it?*

A small utility that checks a list of proxies from `raw.txt` and writes working proxies to `valid.txt` and non-working ones to `invalid.txt`.

This little script originally started as part of another project of mine (which turned out to be a bit too ambitious for me). Rather than letting it go to waste, I decided to turn it into a standalone utility.

There's also an experimental feature that automatically downloads and checks proxy lists from the [iplocate/free-proxy-list](https://github.com/iplocate/free-proxy-list) repository.

One small caveat: although many of the proxies listed there are technically functional, they are blocked by a lot of websites. As they say, there's no such thing as a free lunch.

For now, the only purpose of the `settings.yml` file is to change the application language.

## *What's planned?*

- ~~Add support for loading proxy lists from multiple sources.~~
- ~~Add a Makefile to build the project into a separate directory.~~
- Add support for parsing source files in different formats (the original implementation was designed specifically for [all-proxies.txt](https://raw.githubusercontent.com/iplocate/free-proxy-list/refs/heads/main/all-proxies.txt)). Ideally, users should be able to specify the format of each source. Alternatively, the application could detect the format automatically.
- Add additional settings for configuring the number of threads.

## *Building*

The only requirement is Go 1.27.x or later.

For a quick build, simply run `make` in the terminal. If Make is not installed, you can use the appropriate command for your operating system from the `Makefile`.

After building, a `build` directory will be created in the project root. It will contain the following files:

- `raw.txt`
- `valid.txt`
- `invalid.txt`
- `sources.txt`
- `settings.yml`
- The executable (`proxy_checker` or `proxy_checker.exe`).

**Important:** The Windows build has not been tested yet, so its functionality is not guaranteed.