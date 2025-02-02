FROM golang:1.15-alpine

# from your host - email and telegram settings for authorization
ARG SEND_BOT_TOKEN
ARG SMTP_HOST
ARG SMTP_PORT
ARG FROM_MAIL
ARG MAIL_PASS
ENV SEND_BOT_TOKEN $SEND_BOT_TOKEN
ENV SMTP_HOST $SMTP_HOST
ENV SMTP_PORT $SMTP_PORT
ENV FROM_MAIL $FROM_MAIL
ENV MAIL_PASS $MAIL_PASS

WORKDIR /home
COPY . .
ENV PATH=$PATH:/home/go_sender
RUN chmod ugo+rx go_sender
EXPOSE 9999
CMD [ "go", "run", "." ] go run .
