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

// type UploadVariable struct {
// 	Kind, Data string

// 	// passing in a plain reader
// 	Reader io.Reader
// 	Length int64
// }

// func (ri *RequestInfo) WriteVar(scope, name string, variable UploadVariable) error {

// 	switch variable.Kind {
// 	case TypeBase64:
// 		dec, err := base64.StdEncoding.DecodeString(variable.Data)
// 		if err != nil {
// 			return err
// 		}
// 		variable.Reader = bytes.NewReader(dec)
// 		variable.Length = int64(len(dec))
// 	case TypePlain:
// 		variable.Reader = strings.NewReader(variable.Data)
// 		variable.Length = int64(len(variable.Data))
// 	case TypeFile:
// 		s, _ := os.Stat(variable.Data)
// 		f, err := os.Open(variable.Data)
// 		defer f.Close()
// 		if err != nil {
// 			return err
// 		}
// 		variable.Reader = f
// 		variable.Length = s.Size()
// 	case TypeReader:
// 		// do nothing
// 	default:
// 		return fmt.Errorf("unknown variable kind: %v", variable.Kind)
// 	}

// 	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8889/var", variable.Reader)
// 	if err != nil {
// 		return err
// 	}

// 	q := req.URL.Query()
// 	q.Add("scope", scope)
// 	q.Add("key", name)
// 	q.Add("aid", ri.aid)

// 	// assign encoded query string to http request
// 	req.URL.RawQuery = q.Encode()

// 	req.ContentLength = variable.Length

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	return nil
// }

// func (ri *RequestInfo) ReadVar(scope, name string) (io.ReadCloser, int64, error) {

// 	client := &http.Client{}
// 	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8889/var", nil)
// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	q := req.URL.Query()
// 	q.Add("scope", scope)
// 	q.Add("key", name)
// 	q.Add("aid", ri.aid)
// 	req.URL.RawQuery = q.Encode()

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	lh := resp.Header.Get("content-length")
// 	n, _ := strconv.ParseInt(lh, 10, 64)

// 	return resp.Body, n, nil

// }

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
