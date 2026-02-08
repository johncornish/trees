package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	baseURL := os.Getenv("TREES_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &Client{baseURL: baseURL, http: &http.Client{}}

	switch os.Args[1] {
	case "post-evidence":
		if err := postEvidence(client, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "create-claim":
		if err := createClaim(client, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "link-evidence":
		if err := linkEvidence(client, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "list-claims":
		if err := listClaims(client); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "show-claim":
		if err := showClaim(client, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "list-evidence":
		if err := listEvidence(client); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "show-evidence":
		if err := showEvidence(client, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: trees-cli <command> [args]

Commands:
  post-evidence --file <path> --lines <ref> --commit <hash> [--claim <id>]
      Post a file reference as evidence. Path is resolved to absolute.
      Requires a git commit hash. Optionally link to an existing claim.

  create-claim <content>
      Create a new claim node.

  link-evidence --claim <id> --evidence <id>
      Link an existing evidence node to a claim.

  list-claims
      List all claims.

  show-claim <id>
      Show a claim and its linked evidence.

  list-evidence
      List all evidence nodes.

  show-evidence <id>
      Show an evidence node.

Environment:
  TREES_URL    Server URL (default: http://localhost:8080)
`)
}

type Client struct {
	baseURL string
	http    *http.Client
}

func (c *Client) post(path string, body interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Post(c.baseURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return readJSON(resp)
}

func (c *Client) get(path string) ([]byte, error) {
	resp, err := c.http.Get(c.baseURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func readJSON(resp *http.Response) (map[string]interface{}, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func parseFlag(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func postEvidence(client *Client, args []string) error {
	filePath := parseFlag(args, "--file")
	lineRef := parseFlag(args, "--lines")
	gitCommit := parseFlag(args, "--commit")
	claimID := parseFlag(args, "--claim")

	if filePath == "" || lineRef == "" || gitCommit == "" {
		return fmt.Errorf("usage: post-evidence --file <path> --lines <ref> --commit <hash> [--claim <id>]")
	}

	// Resolve to absolute path
	if !filepath.IsAbs(filePath) {
		abs, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("resolving path: %v", err)
		}
		filePath = abs
	}

	result, err := client.post("/evidence", map[string]string{
		"file_path":  filePath,
		"line_ref":   lineRef,
		"git_commit": gitCommit,
	})
	if err != nil {
		return err
	}

	evID := result["id"].(string)
	fmt.Printf("Created evidence %s\n", evID)
	fmt.Printf("  file: %s\n", result["file_path"])
	fmt.Printf("  lines: %s\n", result["line_ref"])
	fmt.Printf("  commit: %s\n", result["git_commit"])

	// Optionally link to claim
	if claimID != "" {
		_, err := client.post("/claims/"+claimID+"/evidence", map[string]string{
			"evidence_id": evID,
		})
		if err != nil {
			return fmt.Errorf("linking to claim: %v", err)
		}
		fmt.Printf("  linked to claim: %s\n", claimID)
	}

	return nil
}

func createClaim(client *Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: create-claim <content>")
	}
	content := strings.Join(args, " ")

	result, err := client.post("/claims", map[string]string{
		"content": content,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Created claim %s\n", result["id"])
	fmt.Printf("  content: %s\n", result["content"])
	return nil
}

func linkEvidence(client *Client, args []string) error {
	claimID := parseFlag(args, "--claim")
	evidenceID := parseFlag(args, "--evidence")

	if claimID == "" || evidenceID == "" {
		return fmt.Errorf("usage: link-evidence --claim <id> --evidence <id>")
	}

	_, err := client.post("/claims/"+claimID+"/evidence", map[string]string{
		"evidence_id": evidenceID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Linked evidence %s to claim %s\n", evidenceID, claimID)
	return nil
}

func listClaims(client *Client) error {
	body, err := client.get("/claims")
	if err != nil {
		return err
	}

	var claims []map[string]interface{}
	if err := json.Unmarshal(body, &claims); err != nil {
		return err
	}

	if len(claims) == 0 {
		fmt.Println("No claims.")
		return nil
	}

	for _, c := range claims {
		fmt.Printf("%s  %s\n", c["id"], c["content"])
	}
	return nil
}

func showClaim(client *Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: show-claim <id>")
	}

	body, err := client.get("/claims/" + args[0])
	if err != nil {
		return err
	}

	var claim map[string]interface{}
	if err := json.Unmarshal(body, &claim); err != nil {
		return err
	}

	fmt.Printf("Claim: %s\n", claim["id"])
	fmt.Printf("  content: %s\n", claim["content"])
	fmt.Printf("  created: %s\n", claim["created_at"])

	if evidence, ok := claim["evidence"].([]interface{}); ok && len(evidence) > 0 {
		fmt.Printf("  evidence (%d):\n", len(evidence))
		for _, e := range evidence {
			ev := e.(map[string]interface{})
			status := "VALID"
			if valid, ok := ev["valid"].(bool); ok && !valid {
				status = "INVALID"
			}
			fmt.Printf("    [%s] %s  %s  %s  @%s\n", status, ev["id"], ev["file_path"], ev["line_ref"], ev["git_commit"])
		}
	} else {
		fmt.Println("  evidence: (none)")
	}
	return nil
}

func listEvidence(client *Client) error {
	body, err := client.get("/evidence")
	if err != nil {
		return err
	}

	var evidence []map[string]interface{}
	if err := json.Unmarshal(body, &evidence); err != nil {
		return err
	}

	if len(evidence) == 0 {
		fmt.Println("No evidence.")
		return nil
	}

	for _, e := range evidence {
		fmt.Printf("%s  %s  %s  @%s\n", e["id"], e["file_path"], e["line_ref"], e["git_commit"])
	}
	return nil
}

func showEvidence(client *Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: show-evidence <id>")
	}

	body, err := client.get("/evidence/" + args[0])
	if err != nil {
		return err
	}

	var ev map[string]interface{}
	if err := json.Unmarshal(body, &ev); err != nil {
		return err
	}

	fmt.Printf("Evidence: %s\n", ev["id"])
	fmt.Printf("  file: %s\n", ev["file_path"])
	fmt.Printf("  lines: %s\n", ev["line_ref"])
	fmt.Printf("  commit: %s\n", ev["git_commit"])
	if valid, ok := ev["valid"].(bool); ok {
		if valid {
			fmt.Println("  status: VALID")
		} else {
			fmt.Println("  status: INVALID (file changed since commit)")
		}
	}
	fmt.Printf("  created: %s\n", ev["created_at"])
	return nil
}
