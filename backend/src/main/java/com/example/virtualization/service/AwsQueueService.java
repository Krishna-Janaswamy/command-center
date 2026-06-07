package com.example.virtualization.service;

import jakarta.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import software.amazon.awssdk.services.sqs.SqsClient;
import software.amazon.awssdk.services.sqs.model.CreateQueueRequest;
import software.amazon.awssdk.services.sqs.model.DeleteMessageRequest;
import software.amazon.awssdk.services.sqs.model.GetQueueUrlRequest;
import software.amazon.awssdk.services.sqs.model.Message;
import software.amazon.awssdk.services.sqs.model.ReceiveMessageRequest;
import software.amazon.awssdk.services.sqs.model.SendMessageRequest;

import java.util.Collections;
import java.util.List;

@Service
public class AwsQueueService {

    private final SqsClient sqsClient;

    @Value("${aws.sqs.queue-name:service-virtualization-schedule}")
    private String queueName;

    @Value("${aws.sqs.enabled:false}")
    private boolean enabled;

    private String queueUrl;

    public AwsQueueService(SqsClient sqsClient) {
        this.sqsClient = sqsClient;
    }

    @PostConstruct
    public void init() {
        if (!enabled) {
            return;
        }
        try {
            queueUrl = sqsClient.getQueueUrl(GetQueueUrlRequest.builder().queueName(queueName).build()).queueUrl();
        } catch (Exception e) {
            queueUrl = sqsClient.createQueue(CreateQueueRequest.builder().queueName(queueName).build()).queueUrl();
        }
    }

    public boolean isEnabled() {
        return enabled;
    }

    public void sendMessage(String body) {
        if (!enabled || queueUrl == null) {
            return;
        }
        SendMessageRequest request = SendMessageRequest.builder()
                .queueUrl(queueUrl)
                .messageBody(body)
                .build();
        sqsClient.sendMessage(request);
    }

    public List<Message> receiveMessages() {
        if (!enabled || queueUrl == null) {
            return Collections.emptyList();
        }
        ReceiveMessageRequest request = ReceiveMessageRequest.builder()
                .queueUrl(queueUrl)
                .maxNumberOfMessages(5)
                .waitTimeSeconds(5)
                .build();
        return sqsClient.receiveMessage(request).messages();
    }

    public void deleteMessage(String receiptHandle) {
        if (!enabled || queueUrl == null || receiptHandle == null) {
            return;
        }
        sqsClient.deleteMessage(DeleteMessageRequest.builder().queueUrl(queueUrl).receiptHandle(receiptHandle).build());
    }
}
