package lsp

import (
	"io"

	"github.com/formancehq/numscript/internal/lsp/language_server"
)

type ServerOptions struct {
	In  io.Reader
	Out io.Writer
}

func RunServer(options ServerOptions) {
	language_server.RunServer(language_server.ServerArgs[State]{
		InitialState: initialState,
		Handler:      handle,
		In:           options.In,
		Out:          options.Out,
	})
}
