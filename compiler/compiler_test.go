package compiler

import (
	"io/ioutil"
	"path/filepath"
	"pixie/lexer"
	"pixie/parser"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func Test_CompileExamples(t *testing.T) {
	examplesDir := filepath.Join("..", "examples")
	files, err := ioutil.ReadDir(examplesDir)
	require.NoError(t, err, "failed to read direction")

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".pixie" {
			continue
		}

		filePath := filepath.Join(examplesDir, file.Name())
		t.Run(file.Name(), func(t *testing.T) {
			content, err := ioutil.ReadFile(filePath)
			require.NoError(t, err, "failed to read file")

			l := lexer.New(string(content))
			p := parser.New(l)
			node, err := p.Parse()
			require.NoError(t, err, "failed to parse")

			lua, err := Compile(node)
			require.NoError(t, err, "failed to compile")

			snaps.MatchSnapshot(t, lua)
		})
	}
}

func Test_VariableInvalidTypeAssign(t *testing.T) {
	pixie := `
	str = "hello world"
	str = 123
	`

	l := lexer.New(pixie)
	p := parser.New(l)
	node, err := p.Parse()
	require.NoError(t, err, "failed to parse")

	_, err = Compile(node)
	require.ErrorIs(t, err, ErrInvalidTypeAssign)
}
