package com.example.virtualization.service;

import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import software.amazon.awssdk.services.sqs.model.Message;

import java.time.Instant;
import java.util.UUID;

@Component
public class LambdaQueueProcessor {

    private final AwsQueueService awsQueueService;
    private final AwsStorageService awsStorageService;
    private final JdbcTemplate jdbc;

    public LambdaQueueProcessor(AwsQueueService awsQueueService, AwsStorageService awsStorageService, JdbcTemplate jdbc) {
        this.awsQueueService = awsQueueService;
        this.awsStorageService = awsStorageService;
        this.jdbc = jdbc;
    }

    @Scheduled(fixedDelayString = "${aws.sqs.polling-delay-ms:15000}")
    public void pollAndProcessQueue() {
        if (!awsQueueService.isEnabled()) {
            return;
        }

        for (Message message : awsQueueService.receiveMessages()) {
            try {
                processMessage(message.body());
            } catch (Exception e) {
                System.err.println("[Lambda] Failed to process SQS message: " + e.getMessage());
            } finally {
                awsQueueService.deleteMessage(message.receiptHandle());
            }
        }
    }

    private void processMessage(String body) {
        String eventId = UUID.randomUUID().toString();
        String bodyKey = null;
        String storedBody = body;
        if (awsStorageService.isEnabled()) {
            if (storedBody != null) {
                bodyKey = awsStorageService.upload("scheduled-events/" + eventId + ".json", storedBody);
                storedBody = null;
            }
        }

        String sql = "INSERT INTO requests (id, method, url, baseUrl, endpoint, headers, body, bodyS3Key, status, response, responseS3Key, responseHeaders, isRecorded, category, ownerGroup, source) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)";
        jdbc.update(sql,
                eventId,
                "SQS",
                "",
                "",
                "scheduled-poll",
                "{}",
                storedBody,
                bodyKey,
                200,
                "processed",
                null,
                "{}",
                1,
                "other",
                "system",
                "scheduled-lambda"
        );

        System.out.println("[Lambda] Processed scheduled message " + eventId + ", bodyS3Key=" + bodyKey);
    }
}
