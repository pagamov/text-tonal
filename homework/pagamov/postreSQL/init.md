Установка PostgreSQL на Mac Os
Установка через Homebrew:

brew install postgres
Добавляем в автозапуск при старте системы Mac OS.

brew services start postgresql
Также можно запустить вручную:

pg_ctl -D /usr/local/var/postgres start
Перезагрузка PostgreSQL

brew services restart postgresql
Провекра версии PostgreSQL в HomeBrew:

brew info postgresql



Может понадобиться установить так:

brew install postgresql
Проинициализировать дб по умолчанию:

initdb --locale=C -E UTF-8 /opt/homebrew/var/postgresql@14
Экспортировать переменную:

export PATH=/opt/homebrew/bin:$PATH



Не обязательно
Если автозагрузка не сработает. Можно провести такие манипуляции:

Директория автозагрузки находится здесь:

~/Library/LaunchAgents
Создаем симлинк

ln -sfv /usr/local/opt/postgresql/*.plist ~/Library/LaunchAgents
Добавляем в автозагрузку

launchctl load ~/Library/LaunchAgents/homebrew.mxcl.postgresql.plist



Подключение к PostgreSQL на Mac OS
Подключение на Mac OS немного отличается от Linux. Проверим пользователей:

psql -l


Для подключения указываем Owner из таблицы выше:

sudo psql -U Dream -d postgres


Пример создания нового пользователя и БД
Полная инструкция по созданию пользователей и БД:
https://ploshadka.net/postgresql/

Особенности кодировок БД:
проблем с кодировкой - https://ploshadka.net/postgresql-nekotorye-osobennosti-bazy-dannykh-i-kodirovok-v-nejj/

Правильный пример создания
Сначала подключаемся как описано выше:

sudo psql -U пользователь_важен_регистр -d postgres
Создаем пользователя с паролем:

CREATE USER ploshadka WITH PASSWORD '123456';
Назначаем ему кодировку:

ALTER ROLE ploshadka SET client_encoding TO 'utf8';
Создаем БД для русской локали, чтобы была возможность поиска:

CREATE DATABASE dbname TEMPLATE=template0 ENCODING 'UTF-8' LC_COLLATE 'ru_RU.UTF-8' LC_CTYPE 'ru_RU.UTF-8';
Назначаем этому пользователю новую БД:

GRANT ALL PRIVILEGES ON DATABASE dbname TO ploshadka;
Проверим по таблице все ли верно:

\list



Как подключиться к PostgreSQL на Mac OS
На localhost достаточно создать базу данных и подключиться к ней указав только её имя:

dbname='stocks'
Остальное подхватиться по умолчанию.