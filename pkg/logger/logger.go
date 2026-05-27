package logger

import (
	"context"
	"encoding/json"
	"fmt"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"log/slog"
	"maps"
	"net"
	"os"
	"sync"
	"time"
)

var (
	development = "development"
	production  = "production"
)

// MultiHandler рассылает записи во все вложенные обработчики
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		_ = h.Handle(ctx, r)
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: handlers}
}

type GraylogHandler struct {
	addr  string
	conn  net.Conn
	level slog.Level
	extra map[string]any
	attrs []slog.Attr
	host  string
	mu    sync.Mutex
}

func NewGraylogHandler(addr string, level slog.Level, extra map[string]any) (*GraylogHandler, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to graylog at %s: %w", addr, err)
	}

	return &GraylogHandler{
		addr:  addr,
		conn:  conn,
		level: level,
		extra: extra,
	}, nil
}

func (h *GraylogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *GraylogHandler) reconnect() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.conn != nil {
		_ = h.conn.Close()
	}

	conn, err := net.Dial("udp", h.addr)
	if err != nil {
		return err
	}

	h.conn = conn
	return nil
}

func getHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

func slogLevelToSyslog(l slog.Level) int {
	switch l {
	case slog.LevelDebug:
		return 7
	case slog.LevelInfo:
		return 6
	case slog.LevelWarn:
		return 4
	case slog.LevelError:
		return 3
	default:
		return 6
	}
}
func (h *GraylogHandler) Handle(_ context.Context, r slog.Record) error {
	fields := map[string]any{
		"version":       "1.1",
		"host":          getHostname(),
		"short_message": r.Message,
		"timestamp":     float64(r.Time.UnixNano()) / 1e9,
		"level":         slogLevelToSyslog(r.Level),
	}

	for k, v := range h.extra {
		fields["_"+k] = v
	}

	r.Attrs(func(a slog.Attr) bool {
		v := a.Value.Any()
		switch val := v.(type) {
		case func() time.Time:
			fields["_"+a.Key] = val()
		case time.Time:
			fields["_"+a.Key] = val.Unix() // или .Format(time.RFC3339)
		default:
			fields["_"+a.Key] = val
		}
		return true
	})

	data, err := json.Marshal(fields)
	if err != nil {
		return nil // Не блокируем логирование из-за ошибки маршалинга
	}

	h.mu.Lock()
	conn := h.conn
	h.mu.Unlock()

	if conn == nil {
		return nil // Соединение не установлено, пропускаем
	}

	_, err = conn.Write(data)
	if err != nil {
		// Пробуем переподключиться один раз
		if reconnectErr := h.reconnect(); reconnectErr == nil {
			h.mu.Lock()
			conn = h.conn
			h.mu.Unlock()
			_, _ = conn.Write(data) // Попытка повторной отправки, игнорируем ошибку
		}
		return nil // Не блокируем основное логирование
	}

	return nil
}

func (h *GraylogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newExtra := make(map[string]any, len(h.extra)+len(attrs))
	maps.Copy(newExtra, h.extra)
	for _, a := range attrs {
		newExtra[a.Key] = a.Value.Any()
	}
	return &GraylogHandler{
		addr:  h.addr,
		conn:  h.conn,
		level: h.level,
		extra: newExtra,
	}
}

func (h *GraylogHandler) WithGroup(_ string) slog.Handler {
	return h
}

func (h *GraylogHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.conn == nil {
		return nil
	}

	err := h.conn.Close()
	h.conn = nil
	return err
}

// AsyncHandler оборачивает любой slog.Handler и делает его асинхронным
type AsyncHandler struct {
	handler slog.Handler
	ch      chan slog.Record
	wg      sync.WaitGroup
	closed  bool
	mu      sync.Mutex
}

func NewAsyncHandler(ctx context.Context, handler slog.Handler, bufferSize int) *AsyncHandler {
	if bufferSize <= 0 {
		bufferSize = 10000
	}

	ah := &AsyncHandler{
		handler: handler,
		ch:      make(chan slog.Record, bufferSize),
	}

	go ah.worker(ctx)

	return ah
}

func (ah *AsyncHandler) Handle(ctx context.Context, record slog.Record) error {
	ah.mu.Lock()
	if ah.closed {
		ah.mu.Unlock()
		return nil
	}
	ah.mu.Unlock()

	select {
	case ah.ch <- record:
	default:
		fmt.Println("log buffer full")
	}

	return nil
}

func (ah *AsyncHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return ah.handler.Enabled(ctx, level)
}

// WithAttrs проксирует вызов к базовому handler
func (ah *AsyncHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &AsyncHandler{
		handler: ah.handler.WithAttrs(attrs),
		ch:      ah.ch,
	}
}

// WithGroup проксирует вызов к базовому handler
func (ah *AsyncHandler) WithGroup(name string) slog.Handler {
	return &AsyncHandler{
		handler: ah.handler.WithGroup(name),
		ch:      ah.ch,
	}
}

// worker обрабатывает записи из канала
func (ah *AsyncHandler) worker(ctx context.Context) {
	defer ah.wg.Done()
	ah.wg.Add(1)

	for record := range ah.ch {
		if err := ah.handler.Handle(ctx, record); err != nil {
			return
		}
	}
}

// Close закрывает асинхронный handler и ждет завершения обработки
func (ah *AsyncHandler) Close() {
	ah.mu.Lock()
	if !ah.closed {
		ah.closed = true
		close(ah.ch)
	}
	ah.mu.Unlock()

	ah.wg.Wait()
}

func SetupLogger(ctx context.Context, env string, graylogAddr string) (*slog.Logger, func()) {
	var baseHandler slog.Handler

	switch env {
	case production:
		baseHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	case development:
		baseHandler = setupPrettySlog().Handler()
	default:
		baseHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}

	var handler = baseHandler
	var closeFns []func()

	if graylogAddr != "" {
		gray, err := NewGraylogHandler(graylogAddr, slog.LevelInfo, map[string]any{
			"app": "rabbit-saver",
			"env": env,
		})
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to connect to Graylog at %s: %v\n", graylogAddr, err)
			fmt.Println("   Continuing without Graylog logging...")
		} else {
			handler = NewMultiHandler(baseHandler, gray)
			closeFns = append(closeFns, func() { _ = gray.Close() })
		}
	}

	asyncHandler := NewAsyncHandler(ctx, handler, 10000)
	closeFns = append(closeFns, asyncHandler.Close)

	shutdown := func() {
		for i := len(closeFns) - 1; i >= 0; i-- {
			closeFns[i]()
		}
	}

	return slog.New(handler), shutdown
}

func setupPrettySlog() *slog.Logger {
	opts := PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := opts.NewPrettyHandler(os.Stdout)
	return slog.New(handler)
}

//	func Err(err error) slog.Attr {
//		return slog.Attr{
//			Key:   "error",
//			Value: slog.StringValue(err.Error()),
//		}
//	}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
