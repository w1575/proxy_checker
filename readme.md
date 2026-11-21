# Proxy Checker

[🇬🇧English](#what-is-it)

## Что это такое?

Небольшая утилита, которая проверяет список прокси из файла raw.txt и записывет валидные прокси в файл valid.txt, а не валидные в invalid.txt соответственно.
Этот небольшой скрип появился из моего другого проекта (слишком амбициозного для меня). И чтобы не потерять - я решил вынести его в эту небольшую утилитку.

Так же есть экспериментальная функция - скрипт сам скачивает и проверяет файлы с одного репозитория. https://github.com/iplocate/free-proxy-list
. Небольшой нюанс - многие прокси, которые в нем имеются, несмотря на то, что валидные, забанены чуть ли не везде - бесплатный сыр как говорится.

## В планах:
- добавить возможность загрузки адресов из разных источников
- ~~Makefile для сборки проекта в отдельную директорию~~
- Добавить возможность парсить отдельно исходники в разных форматах (оригинал был расчитан на [`all-proxies.txt`](https://raw.githubusercontent.com/iplocate/free-proxy-list/refs/heads/main/all-proxies.txt?utm_source=chatgpt.com)). Чтобы была возможность указывать в каком формате данный источник. Или... автоматом??

## Сборка
Только одно требование - установлен Golang ни ниже 1.22.
Для быстрой сборки - выполняем команду "make" в терминале. Если утилиты make нет - выполняем команду для своей операционной системы из файла Makefile.
После этого в проекте создастся директория build в которой будут 4 файла - raw.txt,  valid.txt, invalid.txt и proxy_checker(.exe).

## What is it?

A small utility that checks a list of proxies from `raw.txt` and writes valid proxies to `valid.txt` and invalid ones to `invalid.txt`.

This little script originated as part of another project of mine (which turned out to be a bit too ambitious for me). Rather than letting it go to waste, I decided to turn it into a standalone utility.

There's also an experimental feature that automatically downloads and checks proxy lists from this repository: [iplocate/free-proxy-list](https://github.com/iplocate/free-proxy-list).

One small caveat: many of the proxies listed there, despite being technically functional, are blocked by a lot of websites. As they say, there's no such thing as a free lunch.

## What's planned?

- Add support for loading proxy lists from multiple sources.
- ~~Add a Makefile to build the project into a separate directory.~~ 
- Add support for parsing source files in different formats (the original implementation was designed specifically for [`all-proxies.txt`](https://raw.githubusercontent.com/iplocate/free-proxy-list/refs/heads/main/all-proxies.txt?utm_source=chatgpt.com)). Ideally, users should be able to specify the format of each source. Or perhaps detect it automatically?

## Building

The only requirement is **Go 1.22 or later**.

For a quick build, simply run `make` in the terminal. If you don't have Make installed, run the appropriate command for your operating system from the `Makefile`.

After building, a `build` directory will be created in the project root. It will contain four files: `raw.txt`, `valid.txt`, `invalid.txt`, and the executable (`proxy_checker` or `proxy_checker.exe`, depending on your OS).