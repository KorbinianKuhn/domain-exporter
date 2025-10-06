package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/korbiniankuhn/domain-exporter/internal/config"
	"github.com/korbiniankuhn/domain-exporter/internal/domain"
	"github.com/korbiniankuhn/domain-exporter/internal/metrics"
	"github.com/korbiniankuhn/domain-exporter/internal/ns_lookup"
	"github.com/korbiniankuhn/domain-exporter/internal/rdap"
	"github.com/korbiniankuhn/domain-exporter/internal/whois"
)

func panicOnError(message string, err error) {
	if err != nil {
		slog.Error(message, "error", err)
		panic(err)
	}
}

func checkDomain(url string) *domain.DomainInfo {
	info, err := rdap.CheckDomain(url)
	slog.Debug("rdap check", "domain", url, "error", err)
	if err == nil {
		return info
	}

	info, err = whois.CheckDomain(url)
	slog.Debug("whois check", "domain", url, "error", err)
	if err == nil {
		return info
	}

	info, err = ns_lookup.CheckDomain(url)
	slog.Debug("ns_lookup check", "domain", url, "error", err)
	if err == nil {
		return info
	}

	info.CheckMethod = domain.CheckMethodFailed
	slog.Warn("All check methods failed", "domain", url)

	return info
}

func main() {
	// Default logger (will be overwritten during config load)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Load config
	c, err := config.Get()
	panicOnError("Failed to load config", err)

	// Init prometheus collector
	m := metrics.NewMetrics()
	http.Handle("/metrics", m.GetMetricsHandler())
	slog.Info("metrics enabled", "url", "/metrics")

	wg := sync.WaitGroup{}
	check := make(chan struct{})

	// Run check on trigger
	wg.Add(1)
	go func() {
		for range check {
			for _, domain := range c.Domains {
				startedAt := time.Now()
				info := checkDomain(domain)
				duration := time.Since(startedAt).Seconds()
				m.SetDomainStatus(duration, info)
				slog.Info("Domain checked", "domain", domain, "method", info.CheckMethod, "duration", duration, "status", info.Status(), "expiry", info.ExpiryDate)
			}
		}
		wg.Done()
	}()

	// Run check on start
	check <- struct{}{}

	slog.Info("domain-exporter started", "check_interval_in_seconds", c.CheckIntervalInSeconds)

	// Run check on interval
	go func() {
		ticker := time.NewTicker(time.Duration(c.CheckIntervalInSeconds) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			check <- struct{}{}
		}
	}()

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	slog.Info("health check endpoint", "url", "/health")

	// Start http server
	s := http.Server{
		Addr: ":2112",
	}
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panicOnError("failed to start http server", err)
		}
	}()
	slog.Info("http server started", "port", "2112")

	// Wait for termination signal
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)

	<-osSignal
	slog.Info("received termination signal, shutting down")

	close(check)

	// Stop http server
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		panicOnError("failed to shutdown http server", err)
	}

	// Run until shutdown is complete
	wg.Wait()
	slog.Info("domain-exporter gracefully stopped")
}
