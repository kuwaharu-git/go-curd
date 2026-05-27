# go-curd

Go (Gin) + MySQL + HTML/CSS/JavaScript で作成したシンプルな TODO アプリです。

## 起動方法

```bash
docker compose up --build
```

起動後に以下へアクセスしてください。

- App: http://localhost:8080
- MySQL: localhost:3306

## API

- `GET /api/todos`
- `POST /api/todos`
- `PUT /api/todos/:id/toggle`
- `DELETE /api/todos/:id`
