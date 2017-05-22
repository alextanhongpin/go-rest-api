FROM iron/base
WORKDIR /app
# copy binary into image
COPY go_api /app/

# copy the config.json to the folder
COPY config.json /app/
ENTRYPOINT ["./go_api"]