package security

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"time"
	"youtube-market/internal/notifier"
)

var (
	monitoringActive bool
	lastCheckTime    time.Time
)

// StartPostgresLogMonitoring запускает мониторинг логов PostgreSQL на подозрительную активность
func StartPostgresLogMonitoring() {
	if monitoringActive {
		return
	}
	monitoringActive = true

	go func() {
		ticker := time.NewTicker(30 * time.Second) // Проверяем каждые 30 секунд
		defer ticker.Stop()

		for range ticker.C {
			checkPostgresLogs()
		}
	}()
}

// checkPostgresLogs проверяет логи PostgreSQL на подозрительную активность
func checkPostgresLogs() {
	// Проверяем логи через docker compose logs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "compose", "logs", "--tail=50", "postgres")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Если команда не выполнилась, это нормально (может быть не в Docker окружении)
		return
	}

	// Ищем подозрительные паттерны
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	lines := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	// Проверяем последние строки на подозрительную активность
	for _, line := range lines {
		if isSuspiciousActivity(line) {
			notifySecurityAlert(line)
		}
	}
}

// isSuspiciousActivity проверяет, содержит ли строка подозрительную активность
func isSuspiciousActivity(line string) bool {
	lineLower := strings.ToLower(line)
	
	// Паттерны подозрительной активности
	suspiciousPatterns := []string{
		"copy.*from program",
		"from program",
		"exec.*program",
		"pg_read_file",
		"lo_import",
		"base64",
		"bash",
		"curl.*195.24.237.73",
		"wget.*195.24.237.73",
		"malicious",
		"injection",
		"sql injection",
		"hack",
		"exploit",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lineLower, pattern) {
			return true
		}
	}

	return false
}

// notifySecurityAlert отправляет уведомление о подозрительной активности
func notifySecurityAlert(logLine string) {
	// Извлекаем IP адрес из лога, если есть
	ip := extractIP(logLine)
	
	details := map[string]interface{}{
		"log_line": logLine,
		"source":   "postgresql",
		"severity": "critical",
	}
	
	if ip != "" {
		details["suspicious_ip"] = ip
	}

	notifier.NotifySecurityAlert(
		"Обнаружена попытка SQL Injection или подозрительная активность в PostgreSQL",
		details,
	)
}

// CheckRecentSecurityEvents проверяет последние логи на подозрительную активность
func CheckRecentSecurityEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Проверяем последние 200 строк логов
	cmd := exec.CommandContext(ctx, "docker", "compose", "logs", "--tail=200", "postgres")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Если команда не выполнилась, это нормально
		return
	}

	// Ищем подозрительные паттерны
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	foundSuspicious := false
	suspiciousLines := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		if isSuspiciousActivity(line) {
			foundSuspicious = true
			suspiciousLines = append(suspiciousLines, line)
		}
	}

	if foundSuspicious {
		details := map[string]interface{}{
			"source":          "postgresql",
			"severity":        "critical",
			"events_found":    len(suspiciousLines),
			"last_log_lines":   suspiciousLines,
		}

		notifier.NotifySecurityAlert(
			"Обнаружена подозрительная активность в логах PostgreSQL (проверка при старте)",
			details,
		)
	}
}

// extractIP извлекает IP адрес из строки лога
func extractIP(line string) string {
	// Простая проверка на IP адрес
	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.Contains(part, ".") && strings.Count(part, ".") == 3 {
			// Может быть IP адрес
			if !strings.Contains(part, ":") {
				return part
			}
		}
	}
	return ""
}

