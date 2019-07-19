package flows

import (
	"encoding/json"
	"strings"
)

// LegacyWebhookPayload is a template that matches the JSON payload sent by legacy webhooks
var LegacyWebhookPayload = `@(json(object(
  "contact", object("uuid", contact.uuid, "name", contact.name, "urn", contact.urn),
  "flow", run.flow,
  "path", run.path,
  "results", foreach_value(results, extract_object, "category", "category_localized", "created_on", "input", "name", "node_uuid", "value"),
  "run", object("uuid", run.uuid, "created_on", run.created_on),
  "input", if(
    input,
    object(
      "attachments", foreach(input.attachments, attachment_parts),
      "channel", input.channel,
      "created_on", input.created_on,
      "text", input.text,
      "type", input.type,
      "urn", if(
        input.urn,
        object(
          "display", default(format_urn(input.urn), ""),
          "path", urn_parts(input.urn).path,
          "scheme", urn_parts(input.urn).scheme
        ),
        null
      ),
      "uuid", input.uuid
    ),
    null
  ),
  "channel", default(input.channel, null)
)))`

// ExtractResponseBody extracts a JSON body from a webhook call response trace
func ExtractResponseBody(response string) json.RawMessage {
	parts := strings.SplitN(response, "\r\n\r\n", 2)

	// this response doesn't have a body
	if len(parts) != 2 || len(parts[1]) == 0 {
		return nil
	}

	body := []byte(parts[1])

	// check if body is valid JSON and can be returned as is
	if json.Valid(body) {
		return body
	}

	// if not, treat body as text and encode as a JSON string
	asString, _ := json.Marshal(string(body))
	return asString
}
