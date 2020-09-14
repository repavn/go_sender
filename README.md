### Desctiption   
This is service for send text messages by HTTP queries (to messangers email, telegram and etc.) from another executed programm (for example - site backend engines, another services).   
#### Set your authorization settings in your host via environment variables (i use gmail account):   
```console
    export FROM_MAIL="<your-gmail-account-email>@gmail.com"
    export MAIL_PASS="<your-gmail-account-password>"
    export SMTP_HOST="smtp.gmail.com"
    export SMTP_PORT="587"
    export SEND_BOT_TOKEN="<your-telegram-bot-token>"
```   
#### Development runtime in Docker: build image, run container   
Install docker to your host os.   

```console
    docker build -t sender --build-arg SEND_BOT_TOKEN --build-arg SMTP_HOST --build-arg SMTP_PORT --build-arg FROM_MAIL --build-arg MAIL_PASS .
    docker run -p 9999:9999 -d -t sender
```   
#### Production instructions   
TODO...
