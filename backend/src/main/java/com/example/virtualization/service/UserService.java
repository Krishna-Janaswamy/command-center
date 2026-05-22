package com.example.virtualization.service;

import com.example.virtualization.model.User;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;
import jakarta.annotation.PostConstruct;

import javax.crypto.SecretKeyFactory;
import javax.crypto.spec.PBEKeySpec;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.security.spec.InvalidKeySpecException;
import java.util.Base64;
import org.springframework.context.annotation.DependsOn;
import java.util.List;

@Service
@DependsOn("dbService")
public class UserService {
    private final JdbcTemplate jdbc;

    @Value("${admin.default-password:}")
    private String defaultAdminPassword;

    public UserService(JdbcTemplate jdbc) {
        this.jdbc = jdbc;
    }

    @PostConstruct
    public void initDefaultAdmin() {
        User admin = findByUsername("admin");
        if (admin == null) {
            registerUser("admin", "admin@123", "admin@example.com", "Admin", "QED_DEV_OPS");
        } else {
            // Auto-fix role/adGroup for admin
            jdbc.update("UPDATE users SET role = ?, adGroup = ? WHERE username = ?", "Admin", "QED_DEV_OPS", "admin");
        }
    }

    public User registerUser(String username, String rawPassword, String email, String role, String adGroup) {
        String hashedPassword = hashPassword(rawPassword);
        jdbc.update("INSERT INTO users (username, password, email, role, adGroup) VALUES (?, ?, ?, ?, ?)",
                username, hashedPassword, email, role, adGroup);
        return findByUsername(username);
    }

    public User findByUsername(String username) {
        List<User> users = jdbc.query("SELECT * FROM users WHERE username = ?", (rs, rowNum) -> {
            User u = new User();
            u.setUsername(rs.getString("username"));
            u.setPassword(rs.getString("password"));
            u.setEmail(rs.getString("email"));
            u.setRole(rs.getString("role"));
            u.setAdGroup(rs.getString("adGroup"));
            u.setCreatedAt(rs.getString("createdAt"));
            return u;
        }, username);
        return users.isEmpty() ? null : users.get(0);
    }

    public boolean checkPassword(String rawPassword, String storedHash) {
        if (storedHash == null || !storedHash.contains(":")) return false;
        String[] parts = storedHash.split(":");
        byte[] salt = Base64.getDecoder().decode(parts[0]);
        String hashToVerify = hashPassword(rawPassword, salt);
        return storedHash.equals(hashToVerify);
    }

    private String hashPassword(String password) {
        byte[] salt = new byte[16];
        new SecureRandom().nextBytes(salt);
        return hashPassword(password, salt);
    }

    private String hashPassword(String password, byte[] salt) {
        try {
            int iterations = 65536;
            int keyLength = 256;
            PBEKeySpec spec = new PBEKeySpec(password.toCharArray(), salt, iterations, keyLength);
            SecretKeyFactory skf = SecretKeyFactory.getInstance("PBKDF2WithHmacSHA256");
            byte[] hash = skf.generateSecret(spec).getEncoded();
            return Base64.getEncoder().encodeToString(salt) + ":" + Base64.getEncoder().encodeToString(hash);
        } catch (NoSuchAlgorithmException | InvalidKeySpecException e) {
            throw new RuntimeException("Error hashing password", e);
        }
    }
}
