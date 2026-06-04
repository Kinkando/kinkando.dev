# Backend Environment Variables

Copy `.env.example` to `.env` and fill in values.

| Variable | Purpose |
|---|---|
| `POSTGRES_DSN` | PostgreSQL DSN for the app server (usually `sslmode=require`) |
| `POSTGRES_MIGRATION_URL` | PostgreSQL DSN for dbmate (can use `sslmode=disable` for Supabase) |
| `MONGO_URI` | MongoDB Atlas URI |
| `MONGO_DB` | MongoDB database name |
| `FIREBASE_CREDENTIALS` | Firebase service-account JSON (inline, not a file path) |
| `GEMINI_API_KEY` | Google Gemini API key (for AI chat) |
| `LINE_CHANNEL_SECRET` | LINE Messaging API channel secret |
| `LINE_CHANNEL_ACCESS_TOKEN` | LINE Messaging API access token |
| `PORT` | HTTP listen port (default 8080) |
| `APP_ENV` | Application environment |
