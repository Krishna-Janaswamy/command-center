package com.example.virtualization.service;

import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.util.UUID;

@Component
public class SqsScheduledPublisher {

    private final AwsQueueService awsQueueService;

    public SqsScheduledPublisher(AwsQueueService awsQueueService) {
        this.awsQueueService = awsQueueService;
    }

    @Scheduled(cron = "${aws.sqs.schedule-cron:0 0/5 * * * ?}")
    public void publishScheduledMessage() {
        if (!awsQueueService.isEnabled()) {
            return;
        }
        String message = String.format("{\"type\": \"scheduled-poll\", \"id\": \"%s\", \"timestamp\": \"%s\"}",
                UUID.randomUUID().toString(), Instant.now().toString());
        awsQueueService.sendMessage(message);
        System.out.println("[SQS] Published scheduled message to queue: " + message);
    }
}
