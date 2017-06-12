############################################################
# Dockerfile to build spring boot app container images
# Based on Ubuntu
############################################################

FROM gcr.io/ubuntu.base:latest

MAINTAINER LGSC

ARG JAR

COPY $JAR /app/app.war

WORKDIR /app/

EXPOSE  8080

CMD java -Djava.security.egd=file:/dev/./urandom -XX:+PrintFlagsFinal -XX:+PrintGCDetails $JAVA_OPTIONS -jar app.war
