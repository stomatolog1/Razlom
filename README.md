# Razlom

Музыкальное приложение для совместного прослушивания и создания общей волны.

## Запуск приложения

Go-сервер раздаёт интерфейс из `web`, а музыкальный каталог сохраняет в SQLite (`razlom.db` в текущей папке).

```powershell
go run ./cmd
```

Откройте [http://localhost:8080](http://localhost:8080). Музыкальный API поддерживает `GET /api/music` и `POST /api/music`; для добавления трека передайте JSON с полями `name`, `artist`, `genre` и `username` (положительный ID пользователя). Путь к файлу базы можно задать переменной окружения `RAZLOM_DB`.

Остальные API:

- `POST /api/likes` и `GET /api/likes?user_id=1`
- `POST /api/playlists` и `GET /api/playlists?user_id=1` (только публичные плейлисты пользователя)
- `POST /api/reposts` и `GET /api/reposts?user_id=1`

Для likes и reposts тело запроса содержит `id_user` и `id_music`. Для плейлиста — `id_user`, `name` и необязательное `private`. Веб-интерфейс пока подключён только к API треков; его плейлисты остаются локальными в браузере.
