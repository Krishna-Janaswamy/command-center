package com.example.virtualization.service;

import jakarta.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import software.amazon.awssdk.core.sync.RequestBody;
import software.amazon.awssdk.services.s3.S3Client;
import software.amazon.awssdk.services.s3.model.CreateBucketRequest;
import software.amazon.awssdk.services.s3.model.GetObjectRequest;
import software.amazon.awssdk.services.s3.model.HeadBucketRequest;
import software.amazon.awssdk.services.s3.model.NoSuchBucketException;
import software.amazon.awssdk.services.s3.model.PutObjectRequest;

import java.io.InputStream;
import java.nio.charset.StandardCharsets;

@Service
public class AwsStorageService {

    private final S3Client s3Client;

    @Value("${aws.s3.bucket-name:service-virtualization-other}")
    private String bucketName;

    @Value("${aws.storage.enabled:false}")
    private boolean enabled;

    public AwsStorageService(S3Client s3Client) {
        this.s3Client = s3Client;
    }

    @PostConstruct
    public void init() {
        if (!enabled) {
            return;
        }
        try {
            s3Client.headBucket(HeadBucketRequest.builder().bucket(bucketName).build());
        } catch (NoSuchBucketException e) {
            s3Client.createBucket(CreateBucketRequest.builder().bucket(bucketName).build());
        } catch (Exception e) {
            System.err.println("Unable to initialize S3 bucket " + bucketName + ": " + e.getMessage());
        }
    }

    public boolean isEnabled() {
        return enabled;
    }

    public String upload(String key, String payload) {
        if (!enabled || key == null) {
            return null;
        }
        String normalized = payload == null ? "" : payload;
        PutObjectRequest request = PutObjectRequest.builder()
                .bucket(bucketName)
                .key(key)
                .build();
        s3Client.putObject(request, RequestBody.fromString(normalized, StandardCharsets.UTF_8));
        return key;
    }

    public String download(String key) {
        if (!enabled || key == null || key.isBlank()) {
            return null;
        }
        try (InputStream stream = s3Client.getObject(GetObjectRequest.builder().bucket(bucketName).key(key).build())) {
            return new String(stream.readAllBytes(), StandardCharsets.UTF_8);
        } catch (Exception e) {
            System.err.println("Unable to read S3 object " + key + ": " + e.getMessage());
            return null;
        }
    }
}
