# Dockerfile - для использования уже собранного бинарника
FROM alpine:latest

WORKDIR /app

# Копируем уже собранный бинарник (не собираем заново)
COPY bot .

COPY migrations ./migrations

# Делаем бинарник исполняемым
RUN chmod +x bot

CMD ["./bot"]