package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LoggerTestSuite struct {
	suite.Suite
	buf *bytes.Buffer
	log Logger
}

func TestLogger(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

func (s *LoggerTestSuite) SetupTest() {
	s.buf = &bytes.Buffer{}
	s.log = newWithWriter("trace", s.buf)
}

func (s *LoggerTestSuite) lastEntry() map[string]any {
	s.T().Helper()

	lines := bytes.Split(bytes.TrimSpace(s.buf.Bytes()), []byte("\n"))
	entry := map[string]any{}
	s.Require().NoError(json.Unmarshal(lines[len(lines)-1], &entry))

	return entry
}

func (s *LoggerTestSuite) TestInfo() {
	s.Run("should log with its key/value fields", func() {
		s.log.Info("hello", "key", "value", "count", 2)

		entry := s.lastEntry()
		s.Equal("info", entry["level"])
		s.Equal("hello", entry["message"])
		s.Equal("value", entry["key"])
		s.EqualValues(2, entry["count"])
	})

	s.Run("should log without fields", func() {
		s.log.Info("plain")

		s.Equal("plain", s.lastEntry()["message"])
	})

	s.Run("should report a trailing key without value as !BADKEY", func() {
		s.log.Info("odd", "a", 1, "dangling")

		entry := s.lastEntry()
		s.EqualValues(1, entry["a"])
		s.Equal("dangling", entry["!BADKEY"])
	})

	s.Run("should report a non-string key as !BADKEY", func() {
		s.log.Info("bad", 42, "x", "ok", "y")

		entry := s.lastEntry()
		s.EqualValues(42, entry["!BADKEY"])
		s.Equal("y", entry["ok"])
	})
}

func (s *LoggerTestSuite) TestWarn() {
	s.Run("should log with its key/value fields", func() {
		s.log.Warn("hello", "key", "value", "count", 2)

		entry := s.lastEntry()
		s.Equal("warn", entry["level"])
		s.Equal("hello", entry["message"])
		s.Equal("value", entry["key"])
		s.EqualValues(2, entry["count"])
	})
}

func (s *LoggerTestSuite) TestDebug() {
	s.Run("should log with its key/value fields", func() {
		s.log.Debug("hello", "key", "value", "count", 2)

		entry := s.lastEntry()
		s.Equal("debug", entry["level"])
		s.Equal("hello", entry["message"])
		s.Equal("value", entry["key"])
		s.EqualValues(2, entry["count"])
	})
}

func (s *LoggerTestSuite) TestTrace() {
	s.Run("should log with its key/value fields", func() {
		s.log.Trace("hello", "key", "value", "count", 2)

		entry := s.lastEntry()
		s.Equal("trace", entry["level"])
		s.Equal("hello", entry["message"])
		s.Equal("value", entry["key"])
		s.EqualValues(2, entry["count"])
	})
}

func (s *LoggerTestSuite) TestError() {
	s.Run("should include the error and message", func() {
		s.log.Error("boom", errors.New("bad"), "key", "value")

		entry := s.lastEntry()
		s.Equal("error", entry["level"])
		s.Equal("boom", entry["message"])
		s.Equal("bad", entry["error"])
		s.Equal("value", entry["key"])
	})
}

func (s *LoggerTestSuite) TestWith() {
	s.Run("should attach fields to the child without mutating the parent", func() {
		child := s.log.With("request_id", "abc")

		child.Info("child")
		s.Equal("abc", s.lastEntry()["request_id"])

		s.log.Info("parent")
		s.NotContains(s.lastEntry(), "request_id")
	})
}

func (s *LoggerTestSuite) TestSetGlobalValue() {
	s.Run("should appear in subsequent output", func() {
		s.log.SetGlobalValue("correlation_id", "xyz")
		s.log.Info("hi")

		s.Equal("xyz", s.lastEntry()["correlation_id"])
	})
}

func (s *LoggerTestSuite) TestWithCtx() {
	s.Run("should round trip the logger through the context", func() {
		ctx := WithCtx(context.Background(), s.log)

		s.Same(s.log, FromCtx(ctx))
	})
}

func (s *LoggerTestSuite) TestFromCtx() {
	s.Run("should return a usable logger for an empty context", func() {
		l := FromCtx(context.Background())

		s.NotNil(l)
		s.NotPanics(func() { l.Info("silent", "k", "v") })
	})
}
