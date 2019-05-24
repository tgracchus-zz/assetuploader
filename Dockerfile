FROM alpine:3.9.4
RUN apk update && apk upgrade && apk add bash && apk add ca-certificates
ADD distributions/assetuploader /bin/assetuploader
ADD build/checkEnv.sh /bin/checkEnv.sh
ADD build/entrypoint.sh /bin/entrypoint.sh
ENV AWS_ACCESS_KEY_ID foo
ENV AWS_SECRET_ACCESS_KEY bar
ENV AWS_REGION eu-west-1
ENV AWS_BUCKET bucket
EXPOSE 8080
ENTRYPOINT [ "/bin/entrypoint.sh" ]