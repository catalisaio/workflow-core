#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  >&2 echo "Usage: $0 <workflow.json> <input.json>"
  exit 2
fi

workflow_path="$1"
input_path="$2"

if [ ! -f "$workflow_path" ]; then
  >&2 echo "Workflow file not found: $workflow_path"
  exit 2
fi

if [ ! -f "$input_path" ]; then
  >&2 echo "Input file not found: $input_path"
  exit 2
fi

if ! command -v docker >/dev/null 2>&1; then
  >&2 echo "docker is required"
  exit 2
fi

if ! command -v jq >/dev/null 2>&1; then
  >&2 echo "jq is required"
  exit 2
fi

if ! command -v python3 >/dev/null 2>&1; then
  >&2 echo "python3 is required"
  exit 2
fi

image="${N8N_DOCKER_IMAGE:-n8nio/n8n:latest}"

docker_net_args=()
if [ -n "${N8N_DOCKER_NETWORK_MODE:-}" ]; then
  docker_net_args+=("--network" "$N8N_DOCKER_NETWORK_MODE")
fi
docker_net_args+=("--add-host" "host.docker.internal:host-gateway")

docker_env_args=()
docker_env_args+=("-e" "N8N_BLOCK_ENV_ACCESS_IN_NODE=false")
if [ -n "${TEST_HTTP_URL:-}" ]; then
  docker_env_args+=("-e" "TEST_HTTP_URL")
fi

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

prepared_workflow="$tmp_dir/workflow.prepared.json"
n8n_home="$tmp_dir/n8n-home"
mkdir -p "$n8n_home"

python3 - "$workflow_path" "$input_path" "$prepared_workflow" <<'PY'
import json
import uuid
import sys

workflow_file, input_file, output_file = sys.argv[1:4]

with open(workflow_file, "r", encoding="utf-8") as f:
    workflow = json.load(f)

with open(input_file, "r", encoding="utf-8") as f:
    input_items = json.load(f)

if not isinstance(input_items, list):
    input_items = []

seed_payload = input_items[0] if input_items and isinstance(input_items[0], dict) else {}

nodes = workflow.setdefault("nodes", [])
connections = workflow.setdefault("connections", {})

manual_trigger = None
for node in nodes:
    if str(node.get("type", "")).lower() == "n8n-nodes-base.manualtrigger":
        manual_trigger = node
        break

if manual_trigger is not None and seed_payload:
    trigger_name = manual_trigger.get("name")
    trigger_position = manual_trigger.get("position") or [0, 0]

    injector_name = "__compat_input__"
    existing_names = {str(n.get("name")) for n in nodes}
    if injector_name in existing_names:
        injector_name = f"__compat_input__{uuid.uuid4().hex[:8]}"

    js_payload = json.dumps(seed_payload, ensure_ascii=False)
    injector_node = {
        "id": str(uuid.uuid4()),
        "name": injector_name,
        "type": "n8n-nodes-base.code",
        "typeVersion": 1,
        "position": [trigger_position[0] + 220, trigger_position[1]],
        "parameters": {
            "jsCode": f"return [{js_payload}];"
        },
    }
    nodes.append(injector_node)

    source = connections.setdefault(trigger_name, {})
    source_main = source.setdefault("main", [[]])
    original_first_output = source_main[0] if source_main and isinstance(source_main[0], list) else []

    source_main[0] = [{"node": injector_name, "type": "main", "index": 0}]

    connections[injector_name] = {
        "main": [original_first_output]
    }

with open(output_file, "w", encoding="utf-8") as f:
    json.dump(workflow, f)
PY

docker run --rm \
  "${docker_net_args[@]}" \
  "${docker_env_args[@]}" \
  -v "$n8n_home:/home/node/.n8n" \
  -v "$prepared_workflow:/data/workflow.json:ro" \
  "$image" \
  import:workflow --input=/data/workflow.json >/dev/null

workflow_ids="$(docker run --rm "${docker_net_args[@]}" "${docker_env_args[@]}" -v "$n8n_home:/home/node/.n8n" "$image" list:workflow --onlyId 2>/dev/null || true)"
set -- $workflow_ids
workflow_id="${1:-}"

if [ -z "$workflow_id" ]; then
  >&2 echo "Failed to resolve imported workflow id"
  exit 1
fi

set +e
execution_json="$(docker run --rm "${docker_net_args[@]}" "${docker_env_args[@]}" -v "$n8n_home:/home/node/.n8n" "$image" execute --id="$workflow_id" --rawOutput 2>&1)"
execute_status=$?
set -e

if [ "$execute_status" -ne 0 ]; then
  >&2 echo "n8n execute failed:"
  >&2 echo "$execution_json"
  exit "$execute_status"
fi

execution_payload="$(printf '%s' "$execution_json" | python3 -c '
import json
import sys

raw = sys.stdin.read()
start = raw.find("{")
if start == -1:
    raise SystemExit("No JSON object found in n8n output")

decoder = json.JSONDecoder()
obj, _ = decoder.raw_decode(raw[start:])
print(json.dumps(obj))
')"

printf '%s' "$execution_payload" | jq -c '
  .data.resultData as $rd
  | ($rd.lastNodeExecuted // "") as $last
  | if $last == "" then
      []
    else
      ($rd.runData[$last][-1].data.main[0] // [])
      | map(.json // .)
    end
'
