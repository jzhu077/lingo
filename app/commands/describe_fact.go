package commands

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/codelingo/lingo/service"
	"github.com/codelingo/lingo/service/server"
	"github.com/juju/errors"
)

func init() {
	register(&cli.Command{
		Hidden: true,
		Name:   "describe-fact",
		Usage:  "Describe a fact belonging to a given lexicon.",
		Action: describeFactAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  util.FormatFlg.String(),
				Usage: "The format for the output. Can be listed (default) or \"json\" encoded.",
			},
			cli.StringFlag{
				Name:  util.OutputFlg.String(),
				Usage: "A filepath to output description to. If the flag is not set, outputs to cli.",
			},
			cli.StringFlag{
				Name:  util.VersionFlg.String(),
				Usage: "The version of the lexicon containing the fact. Leave empty for the latest version.",
			},
		},
	}, false, false, versionRq)
}

func describeFactAction(ctx *cli.Context) {
	err := describeFact(ctx)
	if err != nil {
		util.FatalOSErr(err)
		return
	}
}

func describeFact(ctx *cli.Context) error {
	svc, err := service.New()
	if err != nil {
		return errors.Trace(err)
	}

	var owner, name, lexicon, fact string
	if len(ctx.Args()) > 0 {
		lexicon = ctx.Args()[0]
	}

	if args := strings.Split(lexicon, "/"); len(args) == 3 {
		owner = args[0]
		name = args[1]
		fact = args[2]
	} else {
		return errors.New("Please specify a properly namespaced fact, ie,\nlingo describe-fact codelingo/go/func_decl")
	}

	description, err := svc.DescribeFact(owner, name, ctx.String("version"), fact)
	if err != nil {
		return errors.Trace(err)
	}

	byt := getDescriptionFormat(ctx.String("format"), *description)

	err = outputBytes(ctx.String("output"), byt)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

// TODO(BlakeMScurr) Refactor this and getFormat (from list_lexicons)
// and getFactFormat (from list_facts) which have very similar logic
func getDescriptionFormat(format string, output server.DescribeFactResponse) []byte {
	var content []byte
	switch format {
	case "json":
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(output)
		content = buf.Bytes()
	default:
		content = []byte(formatDescription(output))
	}
	return content
}

func formatDescription(description server.DescribeFactResponse) string {
	// TODO(BlakeMScurr) use a string builder and optimise this
	ret := "Description:\n\t"
	ret += description.Description
	ret += "\nExamples:\n\t"
	ret += description.Examples
	ret += "\nProperties:\n"

	for _, property := range description.Properties {
		ret += "\t" + property.Name + ": " + property.Description + "\n"
	}

	return ret
}
