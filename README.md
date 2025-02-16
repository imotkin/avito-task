### Тестовое задание для стажёра Backend-направления (зимняя волна 2025)


Для запуска приложения и базы данных используется Docker Compose:

```sh 
docker compose up -d
```

После этого сервис будет развёртнут на адресе `localhost:8080`. Для изменения парамтеров запуска возможно использвать файл `.env` со следующими параметрами:

**База данных PostgreSQL**
*	User       – имя пользователя
*	Password   – пароль
*	Host       – хост
*	Port       – порт
*	Database   – база данных

**Приложение** 
*	ServerPort – порт для запуска HTTP-сервера


Для базы данных при запуске её контейнера с образом `postgres:17` выполняется миграция данных и добавляются 4 таблицы для хранения данных сервиса:
* users (данные о пользователях: имени, баланса и пароля в виде хэша)
* transfers (данные о переводах монет между пользователями)
* inventory (данные о купленных товарах для пользователей)
* products (данные о доступных в магазине товарах)

Для работы с базой данных PostgreSQL использован драйвер без CGO: [pq](https://github.com/lib/pq), благодаря этому удалось использовать статический бинарный файл, а значит и более компактный образ для контейнера с приложением на Go.

При выполнении работы старался избегать лишних усложнений для кода и сделать его наиболее очевидным для дальнейшей проверки. 

С интеграционными тестами работал в первый раз, поэтому не удалось в полной мере выполнить и протестировать их работу, также добавил некоторое количество юнит-тестов для покрытия.
Из-за недостатка времени не успел дополнить тесты, а также добавить логгирование для всех операций сервиса, проверку правильности его работы выполнял с помощью Postman. Дополнил проектом файлом YAML для линтера golangci-lint с выбранными правилами для него.
