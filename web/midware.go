package web

import (
	"fmt"
	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"
	"net/http"
	"runtime"
)

// RecoveryWare is the recovery middleware which can cover the panic situation.
type RecoveryWare struct {
	printStack bool
	stackAll   bool
	stackSize  int
}

// ServeHTTP implements the Middleware interface, just recover from the panic. Would provide information on the web page
// if in debug mode.
func (m *RecoveryWare) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, next server.Handler) context.Context {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			stack := make([]byte, m.stackSize)
			stack = stack[:runtime.Stack(stack, m.stackAll)]
			msg := fmt.Sprintf("Request: %s \r\n PANIC: %s\n%s", r.URL.String(), err, stack)
			log.Error(msg)
			if m.printStack {
				fmt.Fprintf(w, "PANIC: %s\n%s", err, stack)
			}
		}
	}()

	return next(ctx, w, r)
}

// NewRecoveryWare returns a new recovery middleware. Would log the full stack if enable the printStack.
func NewRecoveryWare(flags ...bool) server.Middleware {
	stackFlags := []bool{false, false}
	for i := range flags {
		if i >= len(stackFlags) {
			break
		}
		stackFlags[i] = flags[i]
	}
	return &RecoveryWare{
		printStack: stackFlags[0],
		stackAll:   stackFlags[1],
		stackSize:  1024 * 8,
	}
}
