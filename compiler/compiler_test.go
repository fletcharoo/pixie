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

func Test_PrimitiveVariable_InvalidTypeAssign(t *testing.T) {
	pixie := `
	s str = "hello world"
	s = 123
	`

	l := lexer.New(pixie)
	p := parser.New(l)
	node, err := p.Parse()
	require.NoError(t, err, "failed to parse")

	_, err = Compile(node)
	require.ErrorIs(t, err, ErrInvalidTypeAssign)
}

func Test_ListVariable_InvalidTypeAssign(t *testing.T) {
	t.Run("primitive_values", func(t *testing.T) {
		pixie := `
		l list[str] = ["hello", "world"]
		l = 123
		`

		l := lexer.New(pixie)
		p := parser.New(l)
		node, err := p.Parse()
		require.NoError(t, err, "failed to parse")

		_, err = Compile(node)
		require.ErrorIs(t, err, ErrInvalidTypeAssign)
	})

	t.Run("list_value", func(t *testing.T) {
		pixie := `
		l list[str] = ["hello", "world"]
		l = [1, 2, 3]
		`

		l := lexer.New(pixie)
		p := parser.New(l)
		node, err := p.Parse()
		require.NoError(t, err, "failed to parse")

		_, err = Compile(node)
		require.ErrorIs(t, err, ErrInvalidTypeAssign)
	})
}

func Test_MapVariable_InvalidTypeAssign(t *testing.T) {
	t.Run("single_value", func(t *testing.T) {
		pixie := `
		m map[str][num] = {"hello": 1}
		m = {"something": false}
		`

		l := lexer.New(pixie)
		p := parser.New(l)
		node, err := p.Parse()
		require.NoError(t, err, "failed to parse")

		_, err = Compile(node)
		require.ErrorIs(t, err, ErrInvalidTypeAssign)
	})
}
