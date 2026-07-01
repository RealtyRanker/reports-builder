# reports-builder

Сервис регулярных отчётов о квартирах. Раз в 5 минут проверяет таблицу `report_user_subscriptions`, и для каждой подписки, у которой подошло время присылать отчёт, выбирает подходящие квартиры из `flats_history`, формирует CSV-файл `results.csv` и отправляет его пользователю через `users-notifier` вместе с сообщением о топ-5 лучших квартир.

В отличие от `flats-analyzer` (мгновенные уведомления по каждой новой квартире), `reports-builder` работает пакетно: копит квартиры за период подписки (5 минут / 1 час / 12 часов / 24 часа / 7 дней / 30 дней) и присылает их одним отчётом.

## Место в архитектуре

```
                         PostgreSQL (flats_history, report_user_subscriptions)
                                  ↑ читает              ↑ пишут подписки
                                  |                      |
                          reports-builder          subscription-handler (/reports)
                                  |
                                  ↓ POST /send-document (results.csv + подпись)
                            users-notifier  →  Telegram
```

## Что делает сервис

1. Каждые `scheduler.poll_interval_seconds` (по умолчанию 300 — 5 минут) опрашивает `report_user_subscriptions`
2. Отбирает подписки, у которых `last_report_sent_at + period_seconds <= NOW()` и `is_active = TRUE`
3. Для каждой такой подписки:
   - выбирает из `flats_history` квартиры, попавшие в базу в промежутке `(last_report_sent_at, NOW()]`, подходящие под фильтры подписки (те же фильтры, что и в `user_subscriptions` — см. `subscription-handler`/`flats-analyzer`)
   - формирует `results.csv` со всеми полями квартиры
   - берёт топ-5 квартир по `flat_score` и составляет подпись к файлу
   - отправляет файл и подпись через `users-notifier` (`POST /send-document`)
   - обновляет `last_report_sent_at`
4. Если за период не нашлось ни одной подходящей квартиры, файл не отправляется, но `last_report_sent_at` всё равно обновляется (чтобы не копить один и тот же пустой период при следующих проверках)
5. Если отправка не удалась (ошибка Telegram/сети), `last_report_sent_at` не обновляется — отчёт будет повторно собран и отправлен на следующей проверке

### Фильтрация квартир

Часть фильтров подписки применяется прямо в SQL-запросе к `flats_history` (диапазоны цены/площади/этажа/кухни/потолков, список комнат, deal_type, region, обязательные посудомойка/кондиционер/дети/животные, минимальный score). Остальные — там, где на уровне SQL было бы громоздко и подвержено ошибкам — применяются в коде (пакет `internal/filter`):

- минимальное место станции метро в рейтинге (при этом `underground_place = 0` — «нет данных» — всегда считается непройденным)
- минимальный уровень ремонта (`design` > `euro` > `cosmetic` > остальное)
- «есть балкон ИЛИ лоджия»
- тип санузла (раздельный/совмещённый)

Отчёты **не поддерживают** индивидуальные параметры скоринга (в отличие от обычных подписок в `user_subscriptions` + `subscription_scoring_params`) — score для топ-5 и фильтра `min_score` берётся из уже посчитанного `flats_history.flat_score`.

## Конфигурация

```yaml
database:
  dsn: "postgres://realty_parser:password@realty-postgres:5432/realty_parser?sslmode=disable"

notifier:
  base_url: "http://users-notifier:8080"

scheduler:
  poll_interval_seconds: 300   # как часто проверять report_user_subscriptions

logging:
  level: "info"
  file_path: "/var/log/reports-builder/app.log"

metrics:
  port: 9096
```

## Создание и отмена подписок на отчёты

Подписки создаются и отменяются через `subscription-handler` командой `/reports` в Telegram — она предлагает тот же мастер фильтров, что и `/subscript`, плюс выбор периодичности (5 минут / 1 час / 12 часов / 24 часа / 7 дней / 30 дней) сразу после выбора аренда/продажа.

Добавить тестовую подписку напрямую в БД (для локальной отладки):

```sql
-- docker exec -it realty-postgres psql -U realty_parser -d realty_parser

INSERT INTO report_user_subscriptions (chat_id, deal_type, region, period_seconds, last_report_sent_at)
VALUES (123456789, 'rent', 1, 300, NOW() - INTERVAL '1 hour');
```

После этого отчёт по чату `123456789` соберётся при следующем тике планировщика (не позже, чем через `poll_interval_seconds`).

## Метрики

Доступны на порту `9096` (на хосте):

```bash
curl http://localhost:9096/metrics
curl http://localhost:9096/healthz
```

| Метрика | Тип | Описание |
|---|---|---|
| `reports_builder_reports_sent_total` | Counter | Успешно отправленных отчётов |
| `reports_builder_reports_failed_total` | Counter | Ошибок сборки/отправки отчёта |
| `reports_builder_flats_in_report` | Histogram | Количество квартир в отправленном отчёте |

---

## Запуск в Docker

Сервис запускается в сети `realty-net`. Перед запуском должны быть подняты PostgreSQL и `users-notifier`.

### Порядок запуска

```bash
# 1. PostgreSQL (миграции применяются вручную через psql, см. realty-parser/migrations)
cd /путь/к/realty-parser && bash psql_setup.sh

# 2. users-notifier
cd /путь/к/users-notifier && bash server_setup.sh

# 3. reports-builder
cd /путь/к/reports-builder && bash server_setup.sh
```

### Управление контейнером

```bash
# Логи
docker logs -f reports-builder

# Перезапустить немедленно (например, чтобы не ждать poll_interval_seconds
# после добавления тестовой подписки — при старте сервис сразу делает первый тик)
docker restart reports-builder

# Остановить
docker stop reports-builder
```
