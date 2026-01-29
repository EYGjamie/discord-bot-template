docker compose up -d --build --no-deps frontend

docker volume rm discord-bot-template_postgres_data

docker compose down && docker compose build --no-cache frontend && docker compose up -d