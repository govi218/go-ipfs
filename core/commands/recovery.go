package commands

import (
	"fmt"
	"io"

	recovery "github.com/Wondertan/go-ipfs-recovery"
	"github.com/Wondertan/go-ipfs-recovery/reedsolomon"
	"github.com/ipfs/go-cid"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"

	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
)

type EncodeResp struct {
	Cid cid.Cid
}

const (
	recoverabilityOptionName = "recoverability"
)

var RecoveryCmd = &cmds.Command{
	Subcommands: map[string]*cmds.Command{
		"encode": encodeCmd,
	},
}

var encodeCmd = &cmds.Command{
	Arguments: []cmds.Argument{
		cmds.StringArg("path", true, false, "The path to the IPFS object(s) to be encoded.").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.IntOption("recoverability", "r", "Recoverability for the DAG to be encoded with.").WithDefault(3),
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			return err
		}

		path := path.New(req.Arguments[0])
		rnd, err := api.ResolveNode(req.Context, path)
		if err != nil {
			return err
		}

		r, _ := req.Options[recoverabilityOptionName].(int)
		enc, err := recovery.EncodeDAG(req.Context, nd.DAG, reedsolomon.NewEncoder(nd.DAG), rnd, r)
		if err != nil {
			return err
		}

		err = api.Pin().Rm(req.Context, path, options.Pin.RmRecursive(true)) // TODO Make deletion and pinning customizable
		if err != nil {
			return err
		}

		return cmds.EmitOnce(re, &EncodeResp{Cid: enc.Cid()})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, resp *EncodeResp) error {
			_, err := fmt.Fprintf(w, "Encoded: %s", resp.Cid.String())
			return err
		}),
	},
	Type: EncodeResp{},
}
