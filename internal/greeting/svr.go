package greeting

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kydance/ziwi/log"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type GreetingArgs struct {
	Name      string   `json:"name"`
	Age       int      `json:"age"`
	IsVip     bool     `json:"is_vip"`
	Languages []string `json:"languages"`
	Metadata  struct {
		Location string `json:"location"`
		Timezone string `json:"timezone"`
	} `json:"metadata"`
}

func NewMCPServer() *server.MCPServer {
	// Create MCP Server with name, version
	svr := server.NewMCPServer("Greeting", "1.0.0")
	log.Infoln("Creating MCP Server successfully.")

	tool := mcp.NewTool(
		"Greeting",
		mcp.WithDescription("Generate a personalized greeting"),

		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the person to greet")),

		mcp.WithNumber("age",
			mcp.Description("Age of the person to greet"),
			mcp.Min(0),
			mcp.Max(124),
		),

		mcp.WithBoolean("is_vip",
			mcp.Description("Whether the person is a VIP"),
			mcp.DefaultBool(false),
		),

		mcp.WithArray("languages",
			mcp.Description("Languages known by the person"),
			mcp.Items(map[string]any{
				"type": "string",
			}),
		),

		mcp.WithObject(
			"metadata",
			mcp.Description("Additional information about the person"),
			mcp.Properties(
				map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "Current lacation",
					},
					"timezone": map[string]any{
						"type":        "string",
						"description": "Timezone",
					},
				}),
		),
	)
	log.Infoln("Creating Tool successfully.")

	svr.AddTool(tool, mcp.NewTypedToolHandler(greetingHandler))

	return svr
}

func greetingHandler(
	_ context.Context,
	req mcp.CallToolRequest,
	args GreetingArgs,
) (*mcp.CallToolResult, error) {
	log.Infof("GreetingHandler: %+v", req)

	if args.Name == "" {
		log.Errorf("Name is required")
		return mcp.NewToolResultError("Name is required"), nil
	}

	// Build a personalized greeting based on the complex arguments
	sb := strings.Builder{}
	sb.WriteString("Hello, ")
	sb.WriteString(args.Name)
	sb.WriteString(".")

	if args.Age > 0 {
		sb.WriteString(" You are ")
		sb.WriteString(strconv.Itoa(args.Age))
		sb.WriteString(" years old.")
	}

	if args.IsVip {
		sb.WriteString(" Welcome back, valued VIP customer!")
	}

	if len(args.Languages) > 0 {
		sb.WriteString(fmt.Sprintf(" You speak %d languages.", len(args.Languages)))
	}

	if args.Metadata.Location != "" {
		sb.WriteString(" I see you're from ")
		sb.WriteString(args.Metadata.Location)
		sb.WriteString(".")

		if args.Metadata.Timezone != "" {
			sb.WriteString(" Your timezone is ")
			sb.WriteString(args.Metadata.Timezone)
			sb.WriteString(".")
		}
	}

	// Return a text result with the greeting message
	res := mcp.NewToolResultText(sb.String())
	return res, nil
}
