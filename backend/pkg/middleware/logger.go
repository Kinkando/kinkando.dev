package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

const maxBodyLog = 4096

// RequestLogger logs method, path, status, latency, IP, and request/response bodies (capped at 4KB).
func RequestLogger(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		var reqBody []byte
		if b := c.Body(); len(b) > 0 && len(b) <= maxBodyLog {
			reqBody = make([]byte, len(b))
			copy(reqBody, b)
		}

		chainErr := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()

		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.IP()),
			zap.Any("headers", transformHeader(c.GetReqHeaders())),
		}
		if qs := c.Request().URI().QueryString(); len(qs) > 0 {
			fields = append(fields, zap.ByteString("query", qs))
		}
		if len(reqBody) > 0 {
			var req any
			json.Unmarshal(reqBody, &req)
			fields = append(fields, zap.Any("request", req))
		}
		if rb := c.Response().Body(); len(rb) > 0 && len(rb) <= maxBodyLog {
			var res any
			json.Unmarshal(rb, &res)
			fields = append(fields, zap.Any("response", res))
		}

		switch {
		case status >= 500:
			log.Error("http", fields...)
		case status >= 400:
			log.Warn("http", fields...)
		default:
			log.Info("http", fields...)
		}

		return chainErr
	}
}

// MCPRequestLogger wraps an MCP tool handler with structured logging (input, output, latency, errors).
func MCPRequestLogger[In, Out any](logger *zap.Logger, name string, fn func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		start := time.Now()
		inJSON, _ := json.Marshal(in)
		result, out, err := fn(ctx, req, in)
		latency := time.Since(start)
		args := []zap.Field{
			zap.String("tool", name),
			zap.Duration("latency", latency),
		}
		if req.GetExtra() != nil {
			args = append(args, zap.Any("headers", transformHeader(req.GetExtra().Header)))
		}
		if len(inJSON) > 0 {
			args = append(args, zap.Any("input", in))
		}
		if err != nil {
			args = append(args, zap.Error(err))
			logger.Error("mcp", args...)
		} else {
			outJSON, _ := json.Marshal(out)
			if len(outJSON) > 0 {
				args = append(args, zap.Any("output", out))
			}
			logger.Info("mcp", args...)
		}
		return result, out, err
	}
}

func transformHeader(headers http.Header) map[string]any {
	header := make(map[string]any)
	for key, values := range headers {
		if key == "Authorization" {
			header[key] = "REDACTED"
			continue
		}
		if len(values) == 1 {
			header[key] = headers.Get(key)
		} else {
			header[key] = values
		}
	}
	return header
}
