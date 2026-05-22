package com.example.virtualization.util;

import com.example.virtualization.model.User;
import com.example.virtualization.service.JwtService;

public class SecurityHelper {

    public static User extractUser(String authorizationHeader, JwtService jwtService) {
        if (authorizationHeader != null && authorizationHeader.startsWith("Bearer ")) {
            String token = authorizationHeader.substring(7);
            return jwtService.verifyToken(token);
        }
        return null;
    }

    public static boolean isAdmin(User user) {
        return user != null;
    }

    public static boolean canWrite(User user) {
        return user != null;
    }

    public static boolean canDelete(User user) {
        return user != null;
    }

    public static String resolveToggleValue(String value) {
        if (value == null) return "off";
        String lower = value.toLowerCase();
        if (lower.equals("true") || lower.equals("on") || lower.equals("yes") || lower.equals("1")) {
            return "on";
        }
        return "off";
    }

    public static String firstNonBlank(String... strings) {
        for (String s : strings) {
            if (s != null && !s.trim().isEmpty()) {
                return s.trim();
            }
        }
        return null;
    }
}
