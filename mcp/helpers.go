package mcp

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerTool[In any, Out any](server *mcp.Server, tool *mcp.Tool, handler func(ctx context.Context, input In) (Out, error)) {
	mcp.AddTool(server, tool, func(ctx context.Context, request *mcp.CallToolRequest, input In) (result *mcp.CallToolResult, output Out, _ error) {
		in, err := bindInput[In](input)
		if err != nil {
			return nil, *new(Out), err
		}

		out, err := handler(ctx, in)
		return nil, out, err
	})
}

func bindInput[In any](input any) (In, error) {
	if i, ok := input.(In); ok {
		return i, nil
	}

	var result In
	b, err := json.Marshal(input)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(b, &result)
	return result, err
}
