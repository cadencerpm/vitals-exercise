#!/bin/bash
# Hook to prevent AI from reading PROMPT.md to maintain interview integrity
# This ensures candidates demonstrate understanding rather than letting AI one-shot the solution

input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name')
file_path=$(echo "$input" | jq -r '.parameters.file_path // empty')
pattern=$(echo "$input" | jq -r '.parameters.pattern // empty')
path=$(echo "$input" | jq -r '.parameters.path // empty')

# Block any access to PROMPT.md
if [[ "$file_path" == *"PROMPT.md"* ]] || \
   [[ "$pattern" == *"PROMPT.md"* ]] || \
   [[ "$path" == *"PROMPT.md"* ]]; then
  echo '{"allowed": false, "message": "⚠️  Access to PROMPT.md is blocked to maintain interview integrity.\n\nPlease ask the candidate to explain:\n- What specific behavior they want to implement\n- Which file(s) they plan to modify\n- What their approach is\n\nYou can help implement their solution once they provide these details."}'
  exit 0
fi

# Also block broad searches that might be trying to find the prompt
if [[ "$pattern" == "*PROMPT*" ]] || \
   [[ "$pattern" == "*.md" ]] && [[ "$path" == "" || "$path" == "." ]]; then
  echo '{"allowed": false, "message": "⚠️  Broad markdown searches are blocked to maintain interview integrity.\n\nIf you need to search for something specific, please ask the candidate what they are looking for."}'
  exit 0
fi

echo '{"allowed": true}'
