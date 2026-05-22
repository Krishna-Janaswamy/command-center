package com.example.virtualization.config;

import com.example.virtualization.model.User;
import com.example.virtualization.service.JwtService;
import com.example.virtualization.util.SecurityHelper;
import jakarta.servlet.*;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;

import java.io.IOException;

@Component
public class ProxyFilter implements Filter {

    private final JwtService jwtService;

    public ProxyFilter(JwtService jwtService) {
        this.jwtService = jwtService;
    }

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain) throws IOException, ServletException {
        HttpServletRequest req = (HttpServletRequest) request;
        HttpServletResponse res = (HttpServletResponse) response;

        String path = req.getRequestURI();

        // Let API and Actuator Health pass through the filter (they might have their own auth logic in controllers)
        if (path.startsWith("/api/") || path.startsWith("/health") || path.startsWith("/actuator/")) {
            chain.doFilter(request, response);
            return;
        }

        // For non-API paths, enforce auth (this is the catch-all proxy)
        String authHeader = req.getHeader("Authorization");
        User user = SecurityHelper.extractUser(authHeader, jwtService);

        if (user == null) {
            res.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
            res.getWriter().write("Unauthorized");
            return;
        }

        // Here we could implement the actual proxying logic (similar to ProxyController)
        // For now, since ProxyController is handling /api/proxy-external, 
        // this filter just acts as a catch-all auth barrier.
        // The real forwarding will be done in another proxy class or controller.

        chain.doFilter(request, response);
    }
}
