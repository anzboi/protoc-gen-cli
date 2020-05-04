package main

import (
	"bytes"
	"os"
	"testing"

	pgs "github.com/lyft/protoc-gen-star"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestCLIPrinter(t *testing.T) {
	req, err := os.Open("./code_generator_request.pb.bin")
	require.NoError(t, err)

	fs := afero.NewMemMapFs()
	res := bytes.Buffer{}

	pgs.Init(
		pgs.ProtocInput(req),
		pgs.ProtocOutput(&res),
		pgs.FileSystem(fs),
	).RegisterModule(CliPrinter()).Render()

}
