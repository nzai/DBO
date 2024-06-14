package command

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/nzai/dbo/v2/schema"
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
	"github.com/urfave/cli/v3"
)

var (
	ErrOutputPathIsNotADir = errors.New("output path is not a dir")
)

//go:embed model.tmpl
var modelTemplate string

func init() {
	Commands = append(Commands, SchemaToModel{}.Command())
}

type SchemaToModel struct {
	input  string
	output string
	module string
}

func (s SchemaToModel) Command() *cli.Command {
	return &cli.Command{
		Name:    "schema_to_model",
		Aliases: []string{"model"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "input",
				Aliases:     []string{"i"},
				Value:       "schema.sql",
				Usage:       "sql or file path",
				Destination: &s.input,
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Value:       "stdout",
				Usage:       "output dir path",
				Destination: &s.output,
			},
			&cli.StringFlag{
				Name:        "module",
				Aliases:     []string{"m"},
				Value:       "github.com/your/project",
				Usage:       "go module name",
				Destination: &s.module,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if s.output != "stdout" {
				stat, err := os.Stat(s.output)
				if err != nil {
					err = os.MkdirAll(s.output, 0755)
					if err != nil {
						return err
					}
				} else {
					if !stat.IsDir() {
						return ErrOutputPathIsNotADir
					}
				}
			}

			t, err := template.New("").Parse(modelTemplate)
			if err != nil {
				return err
			}

			tables, err := s.Parse(s.input)
			if err != nil {
				return err
			}

			for _, table := range tables {
				b := &bytes.Buffer{}
				err = t.Execute(b, gotable{
					Table:  table,
					Module: s.module,
				})
				if err != nil {
					return err
				}

				buffer, err := format.Source(b.Bytes())
				if err != nil {
					return err
				}

				if s.output == "stdout" {
					os.Stdout.Write(buffer)
					continue
				}

				filePath := filepath.Join(s.output, strings.ToLower(table.SingularName)+"_ag.go")
				err = os.WriteFile(filePath, buffer, 0666)
				if err != nil {
					return err
				}

				log.Printf("%s created", filePath)
			}

			return nil
		},
	}
}

func (s SchemaToModel) Parse(input string) ([]*schema.Table, error) {
	sql, isFile, err := s.tryToReadFile(input)
	if err != nil {
		return nil, err
	}

	if isFile {
		input = sql
	}

	return schema.GetParser().ParseCreateTable(input)
}

func (s SchemaToModel) tryToReadFile(input string) (string, bool, error) {
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

type gotable struct {
	Module string
	*schema.Table
}
