package nrfasthttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Response - object Response Http
type Response struct {
	io.Writer
	h int
}

type errWrapper struct {
	err interface{}
}

func (e errWrapper) Error() string {
	return fmt.Sprintf("generic error %v", e.err)
}

func newResponse(statusCode int) http.ResponseWriter {
	return &Response{h: statusCode}
}

func newResponseWithBody(data []byte, statusCode int) http.ResponseWriter {
	return &Response{Writer: bytes.NewBuffer(data), h: statusCode}
}

// Header - retorna informações do header http
func (w *Response) Header() http.Header { return w.Header() }

// WriteHeader - informa status code http
func (w *Response) WriteHeader(h int) { w.h = h }

func (w *Response) String() string { return fmt.Sprintf("[%v] %q", w.h, w.Writer) }

// NewRelicTransaction - key context transaction newrelic
const NewRelicTransaction = "__newrelic_txn_fasthttp__"

// FromContext - newrelic from context
func FromContext(ctx *fasthttp.RequestCtx) *newrelic.Transaction {
	val := ctx.UserValue(NewRelicTransaction)
	if val == nil {
		return nil
	}
	return val.(*newrelic.Transaction)
}

// Middleware - middleware para fasthpp new relic
func Middleware(app *newrelic.Application, f fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

		if nil == app {
			f(ctx)
			return
		}

		txn := app.StartTransaction(string(ctx.Request.Header.Method()) + " " + string(ctx.Request.URI().Path()))

		defer func() {
			e := recover()

			if e != nil {
				switch err := e.(type) {
				case error:
					txn.NoticeError(err)
				default:
					txn.NoticeError(errWrapper{err})
				}
			}
			txn.End()

			if e != nil {
				panic(e)
			}
		}()

		var r http.Request
		if e := fasthttpadaptor.ConvertRequest(ctx, &r, true); e != nil {
			panic(e)
		}

		txn.SetWebRequestHTTP(&r)
		ctx.SetUserValue(NewRelicTransaction, txn)

		f(ctx)

		statusCode := ctx.Response.StatusCode()

		w := newResponse(statusCode)
		txn.SetWebResponse(w)
	}
}

// NewDataStoreSegment - Newrelic datasource
func NewDataStoreSegment(ctx *fasthttp.RequestCtx, tableName, operation string) *newrelic.DatastoreSegment {

	txn := FromContext(ctx)
	if txn == nil {
		return nil
	}

	s := newrelic.DatastoreSegment{
		StartTime:  txn.StartSegmentNow(),
		Product:    newrelic.DatastoreDynamoDB,
		Collection: tableName,
		Operation:  operation,
	}

	return &s
}

// NewSegment - simple segment
func NewSegment(ctx *fasthttp.RequestCtx, name string) *newrelic.Segment {
	txn := FromContext(ctx)
	if txn == nil {
		return nil
	}

	return txn.StartSegment(name)
}
