package mock

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/admin"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/events"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/scenarios"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/validator"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/webhooks"
)

// Config configures a Server. ListenAddr is a host:port pair; the empty
// string defers to the OS to pick a free port (used in tests).
type Config struct {
	ListenAddr string
	WebhookURL string
	Store      *store.Store
	Logger     *slog.Logger
	// StartedAt records when the mock first booted. Defaults to time.Now()
	// inside New if zero.
	StartedAt time.Time
	// SigningKey is the RSA key used to sign access tokens and seed the
	// JWKS endpoint. Required from Phase 2 onward.
	SigningKey *auth.SigningKey
	// IssuerURL is the value of the `iss` claim and the `issuer` field in
	// the RFC 8414 metadata. Defaults to http://localhost:<port>.
	IssuerURL string
	// TokenTTL overrides the default access-token lifetime when non-zero.
	TokenTTL time.Duration
	// SpecValidator validates incoming requests against the embedded
	// OpenAPI document. Optional; when nil, the server skips validation.
	SpecValidator *validator.Validator
	// PaymentTTL overrides the default PENDING -> COMPLETED transition
	// timer when non-zero.
	PaymentTTL time.Duration
	// QRFormat selects the payload encoded inside QR-type payment QR
	// codes (URL = scan-friendly pay-page URL; EMVCo = production-shape
	// MPM payload). Empty falls back to handlers.DefaultQRFormat.
	QRFormat handlers.QRFormat
}

// Server is the embedded mock HTTP server. It is created with New, started
// with Start, and shut down with Shutdown.
type Server struct {
	cfg       Config
	httpSrv   *http.Server
	listener  net.Listener
	startedAt time.Time
	logger    *slog.Logger
	bus       *events.Bus
	engine    *scenarios.Engine
}

// Bus exposes the in-process event bus so external orchestrators (Phase 5
// webhook dispatcher) can subscribe to state-change events.
func (s *Server) Bus() *events.Bus { return s.bus }

// Engine exposes the scenarios engine for handlers that consult it via
// closures rather than per-request middleware.
func (s *Server) Engine() *scenarios.Engine { return s.engine }

// New builds a Server with its router fully populated. It does not bind a
// listener; Start does that.
func New(cfg Config) (*Server, error) {
	if cfg.Store == nil {
		return nil, fmt.Errorf("mock server: store is required")
	}
	if cfg.Logger == nil {
		return nil, fmt.Errorf("mock server: logger is required")
	}
	if cfg.SigningKey == nil {
		return nil, fmt.Errorf("mock server: signing key is required (Phase 2+)")
	}
	if cfg.StartedAt.IsZero() {
		cfg.StartedAt = time.Now().UTC()
	}
	if cfg.IssuerURL == "" {
		cfg.IssuerURL = deriveIssuerURL(cfg.ListenAddr)
	}
	s := &Server{
		cfg:       cfg,
		startedAt: cfg.StartedAt,
		logger:    cfg.Logger,
		bus:       events.New(),
		engine:    scenarios.New(cfg.Store),
	}
	router := s.buildRouter()
	s.httpSrv = &http.Server{
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s, nil
}

// deriveIssuerURL returns http://localhost:<port> derived from the bind
// address. The localhost form is preferred over the bind interface so the
// JWT iss claim matches what the CLI/merchant typically sees.
func deriveIssuerURL(addr string) string {
	addr = strings.TrimSpace(addr)
	switch {
	case addr == "":
		return "http://localhost:8080"
	case strings.HasPrefix(addr, ":"):
		return "http://localhost" + addr
	case strings.HasPrefix(addr, "0.0.0.0"):
		return "http://localhost" + strings.TrimPrefix(addr, "0.0.0.0")
	case strings.HasPrefix(addr, "[::]"):
		return "http://localhost" + strings.TrimPrefix(addr, "[::]")
	default:
		return "http://" + addr
	}
}

// buildRouter wires the public + admin endpoints onto a single chi router.
// Phase 1 contributed health + admin status. Phase 2 adds the OAuth issuer
// (token endpoint, JWKS, RFC 8414 metadata), the whoami handler behind the
// auth middleware, and the admin keys CRUD surface.
func (s *Server) buildRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(s.requestLogger())
	if s.cfg.SpecValidator != nil {
		r.Use(s.cfg.SpecValidator.Middleware)
	}
	r.Use(s.scenarioMiddleware())

	health := &handlers.Health{Store: s.cfg.Store}
	r.Get("/health/alive", health.Alive)
	r.Get("/health/ready", health.Ready)

	// Mock customer-facing pay page. Public (no auth) — a real customer
	// hitting a payment link has no credentials. Transitions reuse the
	// same code path the TTL worker uses; see transitionPayment.
	payPage := &handlers.PayPage{
		Store:  s.cfg.Store,
		Logger: s.logger,
		Transition: func(_ context.Context, paymentID string, target store.PaymentStatus) error {
			return s.transitionPayment(paymentID, target)
		},
	}
	payPage.Mount(r)

	issuer := &auth.Issuer{
		IssuerURL:  s.cfg.IssuerURL,
		SigningKey: s.cfg.SigningKey,
		Store:      s.cfg.Store,
		TTL:        s.cfg.TokenTTL,
		TTLOverride: func(ctx context.Context) (time.Duration, bool) {
			return scenarios.TokenTTLOverride(ctx, s.engine)
		},
		ScopeFilter: func(ctx context.Context, granted []string) []string {
			return scenarios.FilterScopes(ctx, s.engine, granted)
		},
	}
	r.Post("/oauth2/token", issuer.HandleTokenEndpoint)
	jwksHandler := &auth.JWKSHandler{SigningKey: s.cfg.SigningKey}
	metadataHandler := &auth.MetadataHandler{IssuerURL: s.cfg.IssuerURL, ScopesSupported: auth.AllSpecScopes()}
	r.Get("/.well-known/jwks.json", jwksHandler.ServeHTTP)
	r.Get("/.well-known/oauth-authorization-server", metadataHandler.ServeHTTP)

	authMW := &auth.Middleware{SigningKey: s.cfg.SigningKey, Store: s.cfg.Store}
	r.Route("/api/v1", func(ar chi.Router) {
		ar.Use(authMW.Wrap)
		whoami := &handlers.WhoAmI{}
		ar.Get("/whoami", whoami.ServeHTTP)

		balanceH := &handlers.Balance{Store: s.cfg.Store}
		bankAccountsH := &handlers.BankAccounts{Store: s.cfg.Store}
		historyH := &handlers.History{Store: s.cfg.Store}
		paymentGetH := &handlers.PaymentGet{Store: s.cfg.Store}
		createPaymentH := &handlers.CreatePayment{
			Store:     s.cfg.Store,
			Bus:       s.bus,
			TTL:       s.cfg.PaymentTTL,
			IssuerURL: s.cfg.IssuerURL,
			QRFormat:  s.cfg.QRFormat,
			InsufficientFunds: func(ctx context.Context) bool {
				return scenarios.MaybeInsufficientFunds(ctx, s.engine)
			},
			TransitionAtOverride: func(ctx context.Context, defaultAt time.Time) (time.Time, bool) {
				return scenarios.PaymentStuckOverride(ctx, s.engine, defaultAt)
			},
			TransitionToOverride: func(ctx context.Context, defaultStatus string) (string, time.Duration, bool) {
				return scenarios.PaymentTransitionOverride(ctx, s.engine, defaultStatus)
			},
		}
		createPayoutH := &handlers.CreatePayout{Store: s.cfg.Store, Bus: s.bus}
		streamH := &handlers.StreamPayment{Store: s.cfg.Store, Bus: s.bus}

		transactionGetH := &handlers.TransactionGet{Store: s.cfg.Store}

		ar.With(auth.RequireScope("wallet:balance")).Get("/balance", balanceH.ServeHTTP)
		ar.With(auth.RequireScope("wallet:balance")).Get("/bank-accounts", bankAccountsH.ServeHTTP)
		ar.With(auth.RequireScope("transactions:history")).Get("/history", historyH.ServeHTTP)
		ar.With(auth.RequireScope("transactions:status")).Get("/transactions/{reference}", transactionGetH.ServeHTTP)
		ar.With(auth.RequireScope("transactions:status")).Get("/payments/{paymentId}", paymentGetH.ServeHTTP)
		ar.With(auth.RequireScope("transactions:status")).Get("/payments/{paymentId}/stream", streamH.ServeHTTP)
		// Per D-013 + scaffold §3.1, the auth middleware only does a coarse
		// "has at least one of the listed scopes" check; the per-type
		// (QR / CONTACT / LINK) enforcement happens inside the handler.
		ar.With(auth.RequireAnyScope("payments:qr", "payments:contact", "payments:link")).
			Post("/payments", createPaymentH.ServeHTTP)
		ar.With(auth.RequireScope("wallet:withdraw")).Post("/payouts", createPayoutH.ServeHTTP)
	})

	r.Route("/_admin", func(ar chi.Router) {
		ar.Use(localhostOnly)
		statusHandler := &admin.StatusHandler{
			StartedAt:     s.startedAt,
			ListenAddress: s.cfg.ListenAddr,
			WebhookURL:    s.cfg.WebhookURL,
			Store:         s.cfg.Store,
		}
		ar.Get("/status", statusHandler.ServeHTTP)
		keysHandler := &admin.KeysHandler{Store: s.cfg.Store, DefaultMerchantID: store.DefaultMerchantID}
		keysHandler.Mount(ar)
		logsHandler := &admin.LogsHandler{Store: s.cfg.Store}
		logsHandler.Mount(ar)
		scenariosHandler := &admin.ScenariosHandler{
			Store: s.cfg.Store,
			OnEnable: func(name string, _ map[string]string) {
				switch name {
				case scenarios.ScenarioWebhookDeliveryFail:
					scenarios.ResetWebhookFailureCounter()
				case scenarios.ScenarioClientRevokedAfter:
					scenarios.ScheduleClientRevoke(s.engine, s.cfg.Store, s.logger)
				}
			},
			OnDisable: func(name string) {
				if name == scenarios.ScenarioClientRevokedAfter {
					scenarios.CancelClientRevokes()
				}
			},
		}
		scenariosHandler.Mount(ar)
	})

	return r
}

// transitionPayment is the single point at which a payment moves to a
// terminal status. It mirrors the new status onto the existing
// transaction row (created at payment-create time but hidden from the
// public API until COMPLETED — see store.isVisibleTransaction) and
// publishes an event on the bus so SSE subscribers and the webhook
// dispatcher pick it up. Both the TTL worker and the customer-facing
// pay page invoke this — keeping one implementation prevents the two
// surfaces from drifting on what a transition does.
func (s *Server) transitionPayment(paymentID string, target store.PaymentStatus) error {
	updated, err := s.cfg.Store.UpdatePayment(paymentID, func(p store.Payment) store.Payment {
		p.Status = target
		return p
	})
	if err != nil {
		return fmt.Errorf("update payment %s: %w", paymentID, err)
	}
	if txn, ok, err := s.cfg.Store.FindTransactionBySource(updated.ID); err == nil && ok {
		if err := s.cfg.Store.UpdateTransactionStatus(txn.ID, string(updated.Status)); err != nil {
			s.logger.Warn("transition: update transaction status",
				slog.String("payment_id", updated.ID),
				slog.String("txn_id", txn.ID),
				slog.String("status", string(updated.Status)),
				slog.String("err", err.Error()),
			)
		}
	} else if err != nil {
		s.logger.Warn("transition: find transaction by source",
			slog.String("payment_id", updated.ID),
			slog.String("err", err.Error()),
		)
	}
	s.bus.Publish(events.Event{
		Resource:   "payment",
		ResourceID: updated.ID,
		Status:     string(updated.Status),
		Previous:   string(store.PaymentPending),
		MerchantID: updated.MerchantID,
		OccurredAt: time.Now().UTC(),
	})
	return nil
}

// scenarioMiddleware composes the pre-handler scenarios (latency_injection,
// random_5xx, rate_limit_burst). The FiredHolder is attached upstream by
// the request logger so this layer just appends to it.
func (s *Server) scenarioMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Rate limit first — cheap, drops the request before latency.
			clientID := ""
			if p, ok := auth.PrincipalFromContext(r.Context()); ok {
				clientID = p.ClientID
			}
			if retry, fired := scenarios.MaybeRateLimit(r.Context(), s.engine, clientID, r.URL.Path); fired {
				secs := max(int(retry.Seconds()), 1)
				w.Header().Set("Retry-After", strconv.Itoa(secs))
				problemdetails.WriteNew(w, http.StatusTooManyRequests, problemdetails.TypeRateLimited, "rate limit exceeded; retry after %s", retry)
				return
			}
			if status, fired := scenarios.MaybeRandom5xx(r.Context(), s.engine, r.URL.Path); fired {
				problemdetails.WriteNew(w, status, problemdetails.TypeInternalError, "scenario random_5xx fired (%d)", status)
				return
			}
			scenarios.ApplyLatency(r.Context(), s.engine, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

// requestLogger returns a chi middleware that emits one slog line per
// request, matching the schema in scaffold §11, and persists a structured
// entry to the BoltDB request_log bucket so `swipe logs tail` / `show`
// can recover the data later (D-021 groundwork).
func (s *Server) requestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			holder := &auth.PrincipalHolder{}
			firedHolder := &scenarios.FiredHolder{}
			ctx := auth.WithPrincipalHolder(r.Context(), holder)
			ctx = scenarios.WithFiredHolder(ctx, firedHolder)
			r = r.WithContext(ctx)
			start := time.Now()
			next.ServeHTTP(ww, r)
			elapsed := time.Since(start)

			entry := store.RequestLogEntry{
				StartedAt:  start.UTC(),
				Method:     r.Method,
				Path:       r.URL.Path,
				Status:     ww.Status(),
				DurationMS: elapsed.Milliseconds(),
				RemoteAddr: r.RemoteAddr,
				UserAgent:  r.UserAgent(),
			}
			if holder.Set {
				entry.ClientID = holder.Principal.ClientID
				entry.MerchantID = holder.Principal.MerchantID
			}
			for _, e := range firedHolder.Entries() {
				if e.Effect != "" {
					entry.Scenarios = append(entry.Scenarios, e.Name+":"+e.Effect)
				} else {
					entry.Scenarios = append(entry.Scenarios, e.Name)
				}
			}
			persisted, err := s.cfg.Store.AppendRequestLog(entry)
			if err != nil {
				s.logger.Warn("persist request log", slog.String("err", err.Error()))
				persisted.ID = middleware.GetReqID(r.Context())
			}

			s.logger.Info("request",
				slog.String("req", persisted.ID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int64("duration_ms", elapsed.Milliseconds()),
				slog.String("client", entry.ClientID),
				slog.String("merchant", entry.MerchantID),
			)
		})
	}
}

// Bind synchronously opens the TCP listener so the bound address is
// available before Serve runs. Calling it twice is an error.
func (s *Server) Bind() error {
	if s.listener != nil {
		return fmt.Errorf("server: already bound")
	}
	addr := s.cfg.ListenAddr
	if addr == "" {
		addr = ":8080"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	s.listener = ln
	s.cfg.ListenAddr = ln.Addr().String()
	return nil
}

// Serve runs the HTTP server until ctx is done or it errors. It assumes
// Bind has already been called; calling Serve without Bind first returns
// an error rather than racing.
//
// A background TTL worker is started alongside the HTTP server; it
// transitions PENDING payments to COMPLETED past their TransitionAt and
// publishes events on the bus. The worker exits when ctx is cancelled.
func (s *Server) Serve(ctx context.Context) error {
	if s.listener == nil {
		return fmt.Errorf("server: Serve called before Bind")
	}
	workerCtx, cancelWorker := context.WithCancel(ctx)
	go s.runPaymentTTLWorker(workerCtx)
	defer cancelWorker()
	defer s.bus.Close()

	if s.cfg.WebhookURL != "" {
		secret, _ := s.cfg.Store.GetMeta(store.MetaKeyWebhookSecret)
		dispatcher := &webhooks.Dispatcher{
			URL:    s.cfg.WebhookURL,
			Secret: secret,
			Store:  s.cfg.Store,
			Logger: s.logger,
			ShouldFail: func(ctx context.Context) bool {
				return scenarios.ConsumeWebhookFailure(ctx, s.engine)
			},
		}
		go func() { _ = dispatcher.Run(workerCtx, s.bus) }()
	}

	errCh := make(chan error, 1)
	go func() {
		err := s.httpSrv.Serve(s.listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpSrv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		<-errCh
		return nil
	case err := <-errCh:
		return err
	}
}

// Start is the convenience that runs Bind followed by Serve. It is the
// production lifecycle for the CLI; tests typically call Bind + Serve
// separately so they can read the bound address before serving.
func (s *Server) Start(ctx context.Context) error {
	if err := s.Bind(); err != nil {
		return err
	}
	return s.Serve(ctx)
}

// Listener returns the bound listener. Only valid after Bind has returned.
func (s *Server) Listener() net.Listener { return s.listener }

// Shutdown is a convenience that mirrors http.Server.Shutdown. Production
// code typically uses ctx cancellation on Start instead.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpSrv == nil {
		return nil
	}
	return s.httpSrv.Shutdown(ctx)
}

// localhostOnly rejects any request that did not arrive on a loopback
// connection. Per scaffold §5.8 the /_admin/* surface is unauthenticated
// and the access boundary is "is the caller on this host?".
func localhostOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLoopback(r) {
			http.Error(w, `{"type":"FORBIDDEN","detail":"_admin is localhost-only"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLoopback(r *http.Request) bool {
	host := r.RemoteAddr
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		// Hostnames like "localhost" should also pass.
		return strings.EqualFold(host, "localhost")
	}
	return ip.IsLoopback()
}
