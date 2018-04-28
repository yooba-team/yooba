package coreunix

import (
	"context"

	core "github.com/yooba-team/yooba/yooipfs/core"
	path "github.com/yooba-team/yooba/yooipfs/path"
	resolver "github.com/yooba-team/yooba/yooipfs/path/resolver"
	uio "github.com/yooba-team/yooba/yooipfs/unixfs/io"
)

func Cat(ctx context.Context, n *core.IpfsNode, pstr string) (uio.DagReader, error) {
	r := &resolver.Resolver{
		DAG:         n.DAG,
		ResolveOnce: uio.ResolveUnixfsOnce,
	}

	dagNode, err := core.Resolve(ctx, n.Namesys, r, path.Path(pstr))
	if err != nil {
		return nil, err
	}

	return uio.NewDagReader(ctx, dagNode, n.DAG)
}
