package main

import (
	"context"
	"dota-duel-backend/internal/duel"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	origin := os.Getenv("DUEL_ALLOWED_ORIGIN")
	if _, e := duel.ParseAllowedOrigins(origin); e != nil {
		log.Fatal("invalid DUEL_ALLOWED_ORIGIN configuration")
	}
	dsn := os.Getenv("DUEL_MYSQL_DSN")
	if file := os.Getenv("DUEL_MYSQL_DSN_FILE"); file != "" {
		b, e := os.ReadFile(file)
		if e != nil {
			log.Fatal("cannot read dedicated database secret file")
		}
		dsn = strings.TrimSpace(string(b))
	}
	store, e := duel.OpenMySQL(dsn)
	if e != nil {
		log.Fatal("dedicated MySQL configuration required")
	}
	defer store.DB.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	check, cancel := context.WithTimeout(ctx, 5*time.Second)
	e = store.Ping(check)
	cancel()
	if e != nil {
		log.Fatal("dedicated database connection failed")
	}
	addr := os.Getenv("DUEL_LISTEN")
	if addr == "" {
		addr = "127.0.0.1:18082"
	}
	h := &duel.Handler{Service: duel.NewService(store), Origin: origin, TrustedProxy: os.Getenv("DUEL_TRUSTED_PROXY_IP")}
	server := &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 8192}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				work, cancel := context.WithTimeout(ctx, 5*time.Second)
				if store.Cleanup(work, time.Now().UnixMilli()) != nil {
					log.Print("cleanup deferred: database unavailable")
				}
				cancel()
			}
		}
	}()
	go func() {
		<-ctx.Done()
		end, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(end)
	}()
	log.Print("DOTA DUEL listening on ", addr)
	if e = server.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
		log.Fatal("HTTP listener failed")
	}
}
