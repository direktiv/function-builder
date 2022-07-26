FROM ubuntu:21.04

RUN apt-get update && apt-get install ca-certificates wget openjdk-17-jre git -y

RUN wget https://github.com/karatelabs/karate/releases/download/v1.2.0/karate-1.2.0.jar
RUN mv karate-1.2.0.jar karate.jar

RUN apt-get install dnsutils -y

COPY logging.xml /
COPY karate-config.js /

# ENTRYPOINT ["java", "-DtestURL=10.100.6.10:9191", "-Dlogback.configurationFile=/logging.xml", "-jar", "/karate.jar"]