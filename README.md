### Тестовое задание для стажёра Backend-направления (зимняя волна 2025)

**[Условие задания для сервиса](TASK.md)**

Для работы с проектом возомжно использовать утилиту [Task](https://taskfile.dev/) и добавленные команды в [Taskfile](Taskfile.yml):

1. Запуск приложения и базы данных через Docker Compose
```sh
task start
```

Сервер приложения будет запущен на `localhost:8080`, а сервер PostgreSQL на `localhost:5432`

2. Выполнение сборки проекта (build)
```sh
task build
```

3. Выполнение всех тестов (unit, интеграционные)
```sh 
task test
```

4. Выполнение только unit-тестов
```sh 
task unit-test
```

5. Выполнение только интеграционных тестов
```sh 
task integration-test
```

6. Запуск локально через команду `go`
```sh 
go run ./...
```
---
Для изменения парамтеров запуска возможно использвать файл `.env` со следующими параметрами:

**База данных PostgreSQL**
*	User       – имя пользователя
*	Password   – пароль
*	Host       – хост
*	Port       – порт
*	Database   – база данных
* Logging    – уровень логирования (debug, info, warn, error) 

**Приложение** 
*	ServerPort – порт для запуска HTTP-сервера

Для базы данных при запуске её контейнера с образом `postgres:17` выполняется миграция данных с помощью [goose](https://github.com/pressly/goose) и добавляются 4 таблицы для хранения данных сервиса:
* users (данные о пользователях: имени, баланса и пароля в виде хэша)
* transfers (данные о переводах монет между пользователями)
* inventory (данные о купленных товарах для пользователей)
* products (данные о доступных в магазине товарах)

Для работы с базой данных PostgreSQL использован драйвер без CGO: [pq](https://github.com/lib/pq), благодаря этому удалось использовать статический бинарный файл, а значит и более компактный образ для контейнера с приложением на Go.

При выполнении работы старался избегать лишних усложнений для кода и сделать его наиболее очевидным для дальнейшей проверки. 

Интеграционные тесты были реализованы для двух операций: покупки товара и перевода монет другому пользователю, также добавил некоторое количество юнит-тестов для покрытия.

Для тестирования работы сервиса использовал Postman. 

Дополнил проектом файлом YAML для линтера golangci-lint с выбранными правилами для него.

## Дополнения
Добавлены новые unit-тесты для обработчиков, покрытие **41.6%** больше необходимого порога в 40%.

Результаты для нагрузочного тестирования для 100к запросов и 10 потоков двух обработчиков: для авторизации (`/api/auth`) и получения данных пользователя (`/api/info`). 

**Отклик приложения меньше 50 мс, а RPS выше 1000.**

### 1. Создание Thread Group

<img width="1279" alt="Снимок экрана 2025-02-19 в 22 54 06" src="https://github.com/user-attachments/assets/e7323899-1625-4a0b-89a9-31f31165cede" />

### 2. Таблица запросов для эндпоинта `/api/info`

<img width="1280" alt="Снимок экрана 2025-02-19 в 22 52 19" src="https://github.com/user-attachments/assets/a2e5734f-88e7-4818-a85f-67506f393402" />

### 3. Результат запроса для `/api/info` в виде данных о пользователе в формате JSON

<img width="1277" alt="Снимок экрана 2025-02-19 в 22 53 00" src="https://github.com/user-attachments/assets/d08d48aa-c239-42cc-b2e7-a7e1ca7b04bf" />

### 4. Параметры запроса для `/api/info`

<img width="1278" alt="Снимок экрана 2025-02-19 в 22 53 16" src="https://github.com/user-attachments/assets/29005fcb-56e2-4640-925d-14dbb78164cc" />

### 5. Данные HTTP-хэдера со значенем токена JWT

<img width="1277" alt="Снимок экрана 2025-02-19 в 22 53 30" src="https://github.com/user-attachments/assets/00e8171c-4e9d-4bf3-97fe-ad9ca0418230" />

### 6. Таблица запросов для эндпоинта `/api/auth`

<img width="1280" alt="Снимок экрана 2025-02-19 в 22 57 01" src="https://github.com/user-attachments/assets/9be62f44-23bb-481d-a8af-1ef55abe6562" />

### 7. Параметры запроса для `/api/auth` с телом запроса в виде JSON 

<img width="1279" alt="Снимок экрана 2025-02-19 в 22 57 22" src="https://github.com/user-attachments/assets/143c90db-b7ff-44fc-83d7-f922bbf46bc1" />

### 8. Результат запроса для `/api/auth` в виде токена JWT для пользователя в формате JSON

<img width="1280" alt="Снимок экрана 2025-02-19 в 22 57 34" src="https://github.com/user-attachments/assets/82592e8d-5e53-419d-aa2f-ea53ed1c03a0" />
