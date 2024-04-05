# docker build -t lgbgbl/plagiarism-detection -f Dockerfile .
FROM openjdk:oraclelinux8
RUN mkdir -p /results
WORKDIR /app
COPY plagiarismDetection.jar .
ENTRYPOINT ["java", "-jar", "plagiarismDetection.jar"]