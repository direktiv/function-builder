package apps

import (
	"bytes"
	"fmt"
	"io"
	"os"

	// "encoding/base64"
	// "fmt"
	// "io"
	"net/http"
	// "os"
	// "strconv"
	// "strings"

	"github.com/rs/zerolog"
)

const (
	DirektivActionIDHeader     = "Direktiv-ActionID"
	DirektivErrorCodeHeader    = "Direktiv-ErrorCode"
	DirektivErrorMessageHeader = "Direktiv-ErrorMessage"
	DirektivTmpDir             = "Direktiv-TempDir"

	devMode = "development"
)

type DirektivLogger struct {
	logger zerolog.Logger
}

type RequestInfo struct {
	aid, dir string
	logger   *DirektivLogger
	dl       *DirektivLoggerWriter
}

type DirektivLoggerWriter struct {
	aid string
}

func actionIDFromRequest(r *http.Request) (string, error) {
	aid := r.Header.Get(DirektivActionIDHeader)
	if aid == "" {
		return "", fmt.Errorf("no Direktiv-ActionID header set")
	}
	return aid, nil
}

func RequestinfoFromRequest(req *http.Request) (*RequestInfo, error) {

	aid, err := actionIDFromRequest(req)
	if err != nil {
		return nil, err
	}

	dl := &DirektivLoggerWriter{
		aid: aid,
	}
	cw := consoleWriter(dl)

	return &RequestInfo{
		aid: aid,
		dir: req.Header.Get(DirektivTmpDir),
		dl:  dl,
		logger: &DirektivLogger{
			logger: GetZeroLogger(cw),
		},
	}, nil

}

func timestamp(in interface{}) string {
	return ""
}

func consoleWriter(w io.Writer) zerolog.ConsoleWriter {
	cw := zerolog.ConsoleWriter{Out: w}
	cw.NoColor = true
	cw.FormatTimestamp = timestamp
	cw.FormatLevel = func(i interface{}) string {
		return ""
	}
	return cw
}

func GetZeroLogger(w io.Writer) zerolog.Logger {

	// setup logger
	cw := consoleWriter(os.Stderr)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	var wr io.Writer = cw
	if w != nil {
		wr = io.MultiWriter(cw, w)
	}

	l := zerolog.New(wr).With().Timestamp().Logger()
	return l

}

func (ri *RequestInfo) ActionID() string {
	return ri.aid
}

func (ri *RequestInfo) Logger() *DirektivLogger {
	return ri.logger
}

func (ri *RequestInfo) Dir() string {
	return ri.dir
}

func (ri *RequestInfo) LogWriter() *DirektivLoggerWriter {
	return ri.dl
}

func (dl *DirektivLogger) Errorf(format string, args ...interface{}) {
	txt := fmt.Sprintf(format, args...)
	dl.logger.Error().Msg(txt)
}

func (dl *DirektivLogger) Infof(format string, args ...interface{}) {
	txt := fmt.Sprintf(format, args...)
	dl.logger.Info().Msg(txt)
}

func (dl *DirektivLogger) Debugf(format string, args ...interface{}) {
	txt := fmt.Sprintf(format, args...)
	dl.logger.Debug().Msg(txt)
}

// Write writes log output
func (dl *DirektivLoggerWriter) Write(p []byte) (n int, err error) {

	if dl.aid != devMode {
		_, err = http.Post(fmt.Sprintf("http://localhost:8889/log?aid=%s", dl.aid), "plain/text", bytes.NewBuffer(p))
		return len(p), err
	}

	return len(p), nil

}
