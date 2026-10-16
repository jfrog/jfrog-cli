#!/usr/bin/env bash
# Ported from .jfrog-pipelines/pipelines.yml (JFrog Pipelines is EOL).
#
# Runnable both from the "E2E Dispatch" GitHub Actions workflow and from a
# local shell. To run locally, export the env vars below (same names the
# workflow injects from secrets/inputs) and run this script directly from a
# checkout of jfrog/jfrog-cli.
#
# Required:
#   JENKINS_URL, JENKINS_USER, JENKINS_TOKEN         (jenkins_entplus_rt integration)
#   MASTER_KEY, JFROG_ADMIN_USERNAME,
#   JFROG_ADMIN_PASSWORD, JFROG_ADMIN_TOKEN          (jfrog_cli_tests integration)
#   GITHUB_DISPATCH_TOKEN or JFROG_CLI_GH_TOKEN       (github_dispatch / jfrog_cli_gh integration;
#                                                      fine-grained PAT or GitHub App token —
#                                                      github.jfrog.info rejects classic PATs)
#
# Optional (defaults shown):
#   RT_VERSION=                    SKIP_ENV_SETUP=false      ARTIFACTORY_URL=
#   GITHUB_API_URL=https://github.jfrog.info/api/v3
#   GITHUB_WORKFLOWS_REPO=JFROG/jfrog-cli-workflows          GITHUB_WORKFLOWS_REF=master
#   JFROG_CLI_GITHUB_REPO=jfrog/jfrog-cli                    JFROG_CLI_GITHUB_REF=<current HEAD sha>
#   GHE_ACTIONS_RUNNER=artifactory-dind-amd-scale-set
#   LOGS_TO_KIBANA=true             DEPLOYMENT_SIZING=common
#   MAX_RUN_RETRIES=2               MAX_WAIT_SECONDS=14400
#
# NOT independently verified against a real Jenkins instance: the provision/
# teardown steps call the generic Jenkins remote-build API (crumb + queue
# polling) because JFrog Pipelines' native "Jenkins step type" doesn't expose
# how it itself talks to Jenkins — there was nothing literal to port here.

set -eo pipefail

fail() { echo "ERROR: $1" >&2; exit 1; }

SKIP_ENV_SETUP="${SKIP_ENV_SETUP:-false}"
GITHUB_API_URL="${GITHUB_API_URL:-https://github.jfrog.info/api/v3}"
GITHUB_WORKFLOWS_REPO="${GITHUB_WORKFLOWS_REPO:-JFROG/jfrog-cli-workflows}"
GITHUB_WORKFLOWS_REF="${GITHUB_WORKFLOWS_REF:-master}"
JFROG_CLI_GITHUB_REPO="${JFROG_CLI_GITHUB_REPO:-jfrog/jfrog-cli}"
JFROG_CLI_GITHUB_REF="${JFROG_CLI_GITHUB_REF:-$(git rev-parse HEAD 2>/dev/null || echo "")}"
GHE_ACTIONS_RUNNER="${GHE_ACTIONS_RUNNER:-artifactory-dind-amd-scale-set}"
LOGS_TO_KIBANA="${LOGS_TO_KIBANA:-true}"
DEPLOYMENT_SIZING="${DEPLOYMENT_SIZING:-common}"
MAX_RUN_RETRIES="${MAX_RUN_RETRIES:-2}"
MAX_WAIT_SECONDS="${MAX_WAIT_SECONDS:-14400}"

GITHUB_TOKEN_RESOLVED="${GITHUB_DISPATCH_TOKEN:-${JFROG_CLI_GH_TOKEN:-}}"
[[ -n "${GITHUB_TOKEN_RESOLVED}" ]] || fail "Set GITHUB_DISPATCH_TOKEN (or JFROG_CLI_GH_TOKEN as a fallback)"
[[ -n "${JFROG_CLI_GITHUB_REF}" ]] || fail "Set JFROG_CLI_GITHUB_REF (no git checkout found to default from)"

echo "Starting CLI e2e dispatch"
echo "Using RT_VERSION=${RT_VERSION:-<none>}"

suffix=$(date +%s)
server_name="cli${suffix}"
prefix="${server_name%${suffix}}"
if [[ "${SKIP_ENV_SETUP}" == "true" ]]; then
  [[ -n "${ARTIFACTORY_URL:-}" ]] || fail "SKIP_ENV_SETUP=true requires ARTIFACTORY_URL"
  art_url="${ARTIFACTORY_URL}"
else
  art_url="https://${server_name}.jfrogdev.org"
fi
echo "server_name=${server_name} art_url=${art_url}"

teardown() {
  [[ "${SKIP_ENV_SETUP}" == "true" ]] && return 0
  echo "Tearing down ephemeral Artifactory (${server_name})..."
  local job_path="job/tools/job/platform/job/environment_operate"
  local crumb_json crumb_field crumb
  crumb_json=$(curl -sS -u "${JENKINS_USER}:${JENKINS_TOKEN}" "${JENKINS_URL}/crumbIssuer/api/json" || echo '{}')
  crumb_field=$(echo "$crumb_json" | jq -r '.crumbRequestField // empty')
  crumb=$(echo "$crumb_json" | jq -r '.crumb // empty')
  curl -sS -o /dev/null -u "${JENKINS_USER}:${JENKINS_TOKEN}" \
    ${crumb:+-H "${crumb_field}: ${crumb}"} \
    --data-urlencode "SERVER_NAME=${server_name}" \
    --data-urlencode "ACTION=delete" \
    "${JENKINS_URL}/${job_path}/buildWithParameters" || echo "WARNING: teardown request failed, check Jenkins manually"
}
[[ "${SKIP_ENV_SETUP}" == "true" ]] || trap teardown EXIT

# --- Jenkins step: tools/platform/environment_setup_gen2 ---
# VERIFY: confirm this job path resolves the same way against your Jenkins
# instance as it did through JFrog Pipelines' native Jenkins integration.
if [[ "${SKIP_ENV_SETUP}" != "true" ]]; then
  echo "Provisioning ephemeral Artifactory..."
  job_path="job/tools/job/platform/job/environment_setup_gen2"
  crumb_json=$(curl -sS -u "${JENKINS_USER}:${JENKINS_TOKEN}" "${JENKINS_URL}/crumbIssuer/api/json" || echo '{}')
  crumb_field=$(echo "$crumb_json" | jq -r '.crumbRequestField // empty')
  crumb=$(echo "$crumb_json" | jq -r '.crumb // empty')

  queue_location=$(curl -sS -D - -o /dev/null -u "${JENKINS_USER}:${JENKINS_TOKEN}" \
    ${crumb:+-H "${crumb_field}: ${crumb}"} \
    --data-urlencode "SERVER_NAME=${server_name}" \
    --data-urlencode "DEPLOYMENT_TYPE=onprem" \
    --data-urlencode "ACCOUNT_TYPE=enterprise_plus" \
    --data-urlencode "GROUP=ARTIFACTORY" \
    --data-urlencode "EXPIRY=2d" \
    --data-urlencode "LOGS_TO_KIBANA=${LOGS_TO_KIBANA}" \
    --data-urlencode "DEPLOYMENT_SIZING=${DEPLOYMENT_SIZING}" \
    --data-urlencode "EXTRA_PARAMS=conf_artifactory_unified_version=${RT_VERSION:-} master_key=${MASTER_KEY}" \
    "${JENKINS_URL}/${job_path}/buildWithParameters" \
    | grep -i '^Location:' | tr -d '\r' | awk '{print $2}')
  [[ -n "${queue_location}" ]] || fail "Jenkins did not return a queue item location"

  build_url=""
  for _ in $(seq 1 30); do
    build_url=$(curl -sS -u "${JENKINS_USER}:${JENKINS_TOKEN}" "${queue_location}api/json" | jq -r '.executable.url // empty')
    [[ -n "${build_url}" ]] && break
    sleep 10
  done
  [[ -n "${build_url}" ]] || fail "environment_setup_gen2 never left the queue"
  echo "Building at: ${build_url}"

  while true; do
    build_json=$(curl -sS -u "${JENKINS_USER}:${JENKINS_TOKEN}" "${build_url}api/json")
    [[ "$(echo "$build_json" | jq -r '.building')" == "false" ]] && break
    sleep 15
  done
  result=$(echo "$build_json" | jq -r '.result')
  echo "environment_setup_gen2 result: ${result}"
  [[ "${result}" == "SUCCESS" ]] || fail "environment provisioning failed"
fi

# --- Mint an OAuth token against the (now ready) environment ---
echo "Installing JFrog CLI 2.54.0"
curl -fL https://getcli.jfrog.io/v2 | sh -s 2.54.0
export PATH=$PATH:$HOME/.jfrog
jfrog plugin install access@v7.66.0

support_output=$(jfrog access support-token --url="${art_url}/access" --join-key="${MASTER_KEY}" 2>&1)
support_token=$(echo "${support_output}" | grep -o 'JF_ACCESS_ADMIN_TOKEN=.*' | cut -d'=' -f2-)
[[ -n "${support_token}" ]] || fail "support token empty: ${support_output}"

oauth_response=$(curl -s -w "\nHTTP_CODE:%{http_code}" --location "${art_url}/access/api/v1/oauth/token" \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --header "Authorization: Bearer ${support_token}" \
  --data-urlencode "username=${JFROG_ADMIN_USERNAME}" \
  --data-urlencode 'scope=applied-permissions/admin' \
  --data-urlencode 'expires_in=36000' \
  --data-urlencode 'grant_type=client_credentials' \
  --data-urlencode 'audience=*@*')
oauth_http_code=$(echo "${oauth_response}" | grep "HTTP_CODE:" | cut -d':' -f2)
oauth_body=$(echo "${oauth_response}" | grep -v "HTTP_CODE:")
[[ "${oauth_http_code}" == "200" ]] || fail "oauth HTTP ${oauth_http_code}: ${oauth_body}"

oauth_token=$(echo "${oauth_body}" | jq -r .access_token)
[[ -n "${oauth_token}" && "${oauth_token}" != "null" ]] || fail "could not parse access_token"
echo "OAuth token acquired for ${art_url}"

# --- Dispatch and monitor the GitHub Actions test workflows ---
workflow_files="artifactoryTests.yml goTests.yml npmTests.yml pnpmTests.yml pythonTests.yml mavenTests.yml gradleTests.yml nugetTests.yml conanTests.yml helmTests.yml lifecycleTests.yml accessTests.yml pluginsTests.yml dockerTests.yml podmanTests.yml distributionTests.yml"

repo_code=$(curl -sS -o /tmp/gh_repo.json -w "%{http_code}" \
  -H "Accept: application/vnd.github+json" -H "Authorization: Bearer ${GITHUB_TOKEN_RESOLVED}" \
  -H "X-GitHub-Api-Version: 2022-11-28" "${GITHUB_API_URL}/repos/${GITHUB_WORKFLOWS_REPO}")
[[ "${repo_code}" == "200" ]] || fail "Cannot access ${GITHUB_WORKFLOWS_REPO} (HTTP ${repo_code}): $(cat /tmp/gh_repo.json)"

curl -sS -o /tmp/gh_workflows.json \
  -H "Accept: application/vnd.github+json" -H "Authorization: Bearer ${GITHUB_TOKEN_RESOLVED}" \
  -H "X-GitHub-Api-Version: 2022-11-28" "${GITHUB_API_URL}/repos/${GITHUB_WORKFLOWS_REPO}/actions/workflows"

common_inputs=$(jq -n \
  --arg jfrog_cli_repository "${JFROG_CLI_GITHUB_REPO}" --arg jfrog_cli_ref "${JFROG_CLI_GITHUB_REF}" \
  --arg jfrog_url "${art_url}" --arg jfrog_admin_token "${oauth_token}" --arg runner "${GHE_ACTIONS_RUNNER}" \
  '{jfrog_cli_repository:$jfrog_cli_repository, jfrog_cli_ref:$jfrog_cli_ref, jfrog_url:$jfrog_url, jfrog_admin_token:$jfrog_admin_token, runner:$runner}')

dispatched=""
for wf_file in ${workflow_files}; do
  workflow_id=$(jq -r --arg f "${wf_file}" '.workflows[] | select(.path | endswith($f)) | .id' /tmp/gh_workflows.json | head -1)
  [[ -z "${workflow_id}" || "${workflow_id}" == "null" ]] && { echo "WARNING: ${wf_file} not indexed, skipping"; continue; }

  case "${wf_file}" in
    distributionTests.yml) inputs=$(echo "${common_inputs}" | jq --arg jfrog_user "${JFROG_ADMIN_USERNAME:-admin}" '. + {jfrog_user:$jfrog_user}') ;;
    *) inputs="${common_inputs}" ;;
  esac
  dispatch_body=$(jq -n --arg ref "${GITHUB_WORKFLOWS_REF}" --argjson inputs "${inputs}" '{ref:$ref, inputs:$inputs}')

  http_code=$(curl -sS -o /tmp/gh_dispatch_resp.txt -w "%{http_code}" -X POST \
    -H "Accept: application/vnd.github+json" -H "Authorization: Bearer ${GITHUB_TOKEN_RESOLVED}" \
    -H "X-GitHub-Api-Version: 2022-11-28" -d "${dispatch_body}" \
    "${GITHUB_API_URL}/repos/${GITHUB_WORKFLOWS_REPO}/actions/workflows/${workflow_id}/dispatches")
  [[ "${http_code}" == "204" ]] || fail "Dispatch failed for ${wf_file} (HTTP ${http_code}): $(cat /tmp/gh_dispatch_resp.txt)"
  echo "Dispatched ${wf_file} (id=${workflow_id})"
  dispatched="${dispatched} ${wf_file}:${workflow_id}"
done
[[ -n "${dispatched}" ]] || fail "No workflows were dispatched"

echo "Waiting 30s for runs to appear..."
sleep 30
run_entries=""
for entry in ${dispatched}; do
  wf_file="${entry%%:*}"; wf_id="${entry##*:}"; run_id=""
  for attempt in $(seq 1 20); do
    run_id=$(curl -sS -H "Accept: application/vnd.github+json" -H "Authorization: Bearer ${GITHUB_TOKEN_RESOLVED}" \
      -H "X-GitHub-Api-Version: 2022-11-28" \
      "${GITHUB_API_URL}/repos/${GITHUB_WORKFLOWS_REPO}/actions/workflows/${wf_id}/runs?per_page=1&event=workflow_dispatch" \
      | jq -r '.workflow_runs[0].id // empty')
    [[ -n "${run_id}" && "${run_id}" != "null" ]] && break
    sleep 10
  done
  [[ -n "${run_id}" && "${run_id}" != "null" ]] || fail "Could not resolve run id for ${wf_file}"
  run_entries="${run_entries} ${wf_file}:${run_id}"
done

retries_left=""
for entry in ${run_entries}; do wf_file="${entry%%:*}"; retries_left="${retries_left} ${wf_file}=${MAX_RUN_RETRIES}"; done
get_retries_left() { echo "${retries_left}" | tr ' ' '\n' | awk -F= -v n="$1" '$1==n{print $2; exit}'; }
set_retries_left() { local n="$1" c="$2" r=""; for kv in ${retries_left}; do case "$kv" in "$n="*) r="$r $n=$c";; *) r="$r $kv";; esac; done; retries_left="$r"; }

elapsed=0; interval=60
while [[ ${elapsed} -lt ${MAX_WAIT_SECONDS} ]]; do
  all_done=true; any_failed=false; summary=""
  for entry in ${run_entries}; do
    wf_file="${entry%%:*}"; run_id="${entry##*:}"
    run_json=$(curl -sS -H "Accept: application/vnd.github+json" -H "Authorization: Bearer ${GITHUB_TOKEN_RESOLVED}" \
      -H "X-GitHub-Api-Version: 2022-11-28" "${GITHUB_API_URL}/repos/${GITHUB_WORKFLOWS_REPO}/actions/runs/${run_id}")
    status=$(echo "$run_json" | jq -r '.status // empty')
    conclusion=$(echo "$run_json" | jq -r '.conclusion // empty')
    if [[ "${status}" != "completed" ]]; then
      all_done=false; summary="${summary}${wf_file}:${status};"
    elif [[ "${conclusion}" != "success" ]]; then
      rl=$(get_retries_left "${wf_file}"); rl="${rl:-0}"
      if [[ "${rl}" -gt 0 ]]; then
        rr_code=$(curl -sS -o /dev/null -w "%{http_code}" -X POST \
          -H "Accept: application/vnd.github+json" -H "Authorization: Bearer ${GITHUB_TOKEN_RESOLVED}" \
          -H "X-GitHub-Api-Version: 2022-11-28" "${GITHUB_API_URL}/repos/${GITHUB_WORKFLOWS_REPO}/actions/runs/${run_id}/rerun-failed-jobs")
        if [[ "${rr_code}" == "201" ]]; then
          set_retries_left "${wf_file}" "$((rl-1))"; all_done=false; summary="${summary}${wf_file}:retrying;"
        else
          any_failed=true; summary="${summary}${wf_file}:${conclusion};"
        fi
      else
        any_failed=true; summary="${summary}${wf_file}:${conclusion};"
      fi
    else
      summary="${summary}${wf_file}:success;"
    fi
  done
  if [[ "${all_done}" == "true" ]]; then
    [[ "${any_failed}" == "true" ]] && fail "One or more workflow runs failed. ${summary}"
    echo "All workflow runs completed successfully. ${summary}"
    exit 0
  fi
  sleep "${interval}"; elapsed=$((elapsed + interval))
done
fail "Timed out after ${MAX_WAIT_SECONDS}s. ${summary}"
