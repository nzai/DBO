package command

import (
	"bytes"
	"context"
	_ "embed"
	"go/format"
	"io"
	"os"
	"text/template"

	"github.com/nzai/dbo/v2/schema"
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
	"github.com/urfave/cli/v3"
)

//go:embed entity.tmpl
var entityTemplate string

func init() {
	Commands = append(Commands, SchemeToEntity{}.Command())
}

type SchemeToEntity struct {
	Input  string
	Output string
}

func (s SchemeToEntity) Command() *cli.Command {
	return &cli.Command{
		Name:    "schema_to_entity",
		Aliases: []string{"entity"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "input",
				Aliases:     []string{"i"},
				Value:       "schema.sql",
				Usage:       "sql or file path",
				Destination: &s.Input,
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Value:       "stdout",
				Usage:       "output file path",
				Destination: &s.Output,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			var output io.Writer = os.Stdout
			if s.Output != "stdout" {
				file, err := os.Create(s.Output)
				if err != nil {
					return err
				}
				defer file.Close()

				output = file
			}

			t, err := template.New("").Parse(entityTemplate)
			if err != nil {
				return err
			}

			tables, err := s.Parse(s.Input)
			if err != nil {
				return err
			}

			b := &bytes.Buffer{}
			err = t.Execute(b, tables)
			if err != nil {
				return err
			}

			buffer, err := format.Source(b.Bytes())
			if err != nil {
				return err
			}

			_, err = output.Write(buffer)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func (s SchemeToEntity) Parse(input string) ([]*schema.Table, error) {
	sql, isFile, err := s.tryToReadFile(input)
	if err != nil {
		return nil, err
	}

	if isFile {
		input = sql
	}

	return schema.GetParser().ParseCreateTable(input)
}

func (s SchemeToEntity) tryToReadFile(input string) (string, bool, error) {
	_, err := os.Stat(input)
	if err == nil {
		buffer, err := os.ReadFile(input)
		if err != nil {
			return "", false, err
		}

		return string(buffer), true, nil
	}

	return "", false, nil
}
