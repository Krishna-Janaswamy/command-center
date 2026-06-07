package com.example.virtualization;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class ServiceVirtualizationApplication {
    public static void main(String[] args) {
        SpringApplication.run(ServiceVirtualizationApplication.class, args);
    }
}
