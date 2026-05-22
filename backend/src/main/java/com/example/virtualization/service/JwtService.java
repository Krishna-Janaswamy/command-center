package com.example.virtualization.service;

import com.example.virtualization.model.User;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.security.InvalidKeyException;
import java.security.NoSuchAlgorithmException;
import java.util.Base64;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;

import com.example.virtualization.util.JsonUtil;

@Service
public class JwtService {

    @Value("${jwt.secret:dev-default-secret-key-12345678901234567890}")
    private String jwtSecret;

    @Value("${jwt.expiration:86400000}")
    private long jwtExpiration;

    public String generateToken(User user) {
        long now = System.currentTimeMillis();
        
        Map<String, Object> header = new HashMap<>();
        header.put("alg", "HS256");
        header.put("typ", "JWT");

        Map<String, Object> payload = new HashMap<>();
        payload.put("sub", user.getUsername());
        payload.put("role", user.getRole());
        payload.put("adGroup", user.getAdGroup());
        payload.put("iat", now / 1000);
        payload.put("exp", (now + jwtExpiration) / 1000);

        String headerBase64 = encodeBase64Url(JsonUtil.toJson(header));
        String payloadBase64 = encodeBase64Url(JsonUtil.toJson(payload));

        String signature = sign(headerBase64 + "." + payloadBase64, jwtSecret);

        return headerBase64 + "." + payloadBase64 + "." + signature;
    }

    public User verifyToken(String token) {
        if (token == null || token.isEmpty()) return null;

        String[] parts = token.split("\\.");
        if (parts.length != 3) return null;

        String header = parts[0];
        String payload = parts[1];
        String signature = parts[2];

        String expectedSignature = sign(header + "." + payload, jwtSecret);
        if (!expectedSignature.equals(signature)) {
            return null; // Invalid signature
        }

        String payloadJson = decodeBase64Url(payload);
        Map<String, String> claims = JsonUtil.fromJson(payloadJson);

        long exp = Long.parseLong(String.valueOf(claims.get("exp")));
        if (System.currentTimeMillis() / 1000 > exp) {
            return null; // Expired
        }

        User user = new User();
        user.setUsername(claims.get("sub"));
        user.setRole(claims.get("role"));
        user.setAdGroup(claims.get("adGroup"));

        return user;
    }

    private String sign(String data, String secret) {
        try {
            Mac mac = Mac.getInstance("HmacSHA256");
            SecretKeySpec secretKeySpec = new SecretKeySpec(secret.getBytes(StandardCharsets.UTF_8), "HmacSHA256");
            mac.init(secretKeySpec);
            byte[] signatureBytes = mac.doFinal(data.getBytes(StandardCharsets.UTF_8));
            return encodeBase64Url(signatureBytes);
        } catch (NoSuchAlgorithmException | InvalidKeyException e) {
            throw new RuntimeException("Error signing JWT", e);
        }
    }

    private String encodeBase64Url(String data) {
        return encodeBase64Url(data.getBytes(StandardCharsets.UTF_8));
    }

    private String encodeBase64Url(byte[] bytes) {
        return Base64.getUrlEncoder().withoutPadding().encodeToString(bytes);
    }

    private String decodeBase64Url(String base64Url) {
        return new String(Base64.getUrlDecoder().decode(base64Url), StandardCharsets.UTF_8);
    }
}
