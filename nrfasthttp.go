package nrfasthttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type errWrapper struct {
	err interface{}
}

func (e errWrapper) Error() string {
	return fmt.Sprintf("generic error %v", e.err)
}

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
		if app == nil {
			f(ctx)
			return
		}
		isNil := func(value interface{}) bool {
			if value == nil {
				return true
			}
			if reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil() {
				return true
			}
			if reflect.ValueOf(value).Kind() == reflect.Invalid {
				return true
			}
			return false
		}

		toJSON := func(v interface{}) string {
			data, _ := json.Marshal(v)
			return string(data)
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

		if ctx.Response.StatusCode() > 300 {
			if val := ctx.UserValue("payload"); !isNil(val) {
				txn.NoticeError(errors.New(toJSON(val)))
			}
		}

		w := txn.SetWebResponse(nil)
		w.WriteHeader(ctx.Response.StatusCode())
	}
}

// NewExternalSegment - external segment
func NewExternalSegment(ctx *fasthttp.RequestCtx, request *fasthttp.Request) *newrelic.ExternalSegment {
	txn := FromContext(ctx)
	if txn == nil {
		return nil
	}
	newCTX := fasthttp.RequestCtx{}
	request.CopyTo(&newCTX.Request)
	var r http.Request
	if e := fasthttpadaptor.ConvertRequest(&newCTX, &r, true); e != nil {
		return nil
	}
	return newrelic.StartExternalSegment(txn, &r)
}

// NewSegment - simple segment
func NewSegment(ctx *fasthttp.RequestCtx, name string) *newrelic.Segment {
	txn := FromContext(ctx)
	if txn == nil {
		return nil
	}
	return txn.StartSegment(name)
}

// NewDataStoreSegment - Newrelic datasource
func NewDataStoreSegment(ctx *fasthttp.RequestCtx, tableName, operation string) *newrelic.DatastoreSegment {
	txn := FromContext(ctx)
	if txn == nil {
		return nil
	}
	s := newrelic.DatastoreSegment{
		StartTime:  newrelic.StartSegmentNow(txn),
		Product:    newrelic.DatastoreDynamoDB,
		Collection: tableName,
		Operation:  operation,
	}
	return &s
}

// EndSegment - endSegment
func EndSegment(v interface{}) {
	if s, ok := v.(*newrelic.ExternalSegment); ok {
		s.End()
	}
	if s, ok := v.(*newrelic.DatastoreSegment); ok {
		s.End()
	}
	if s, ok := v.(*newrelic.Segment); ok {
		s.End()
	}
	if s, ok := v.(*newrelic.MessageProducerSegment); ok {
		s.End()
	}
}
