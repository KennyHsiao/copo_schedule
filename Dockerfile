FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV TZ Asia/Shanghai

WORKDIR /app
COPY /schedule_service /app
COPY etc/ /app/etc

CMD ["/app/schedule_service", "-env" , "etc/.env"]