#!/bin/sh

set -eu

coverage_dir="${COVERAGE_DIR:-coverage}"
threshold="${COVERAGE_THRESHOLD:-30}"
services="${SERVICES:-analytics fetcher modules scheduler users}"
summary_file="$coverage_dir/services-coverage.txt"

packages_for_service() {
	case "$1" in
	analytics)
		printf '%s\n' "./internal/analytics/service ./internal/analytics/infra/http"
		;;
	fetcher)
		printf '%s\n' "./internal/response/service ./internal/response/infra/grpc"
		;;
	modules)
		printf '%s\n' "./internal/module/service ./internal/module/infra/http/... ./internal/webhook/service ./internal/webhook/infra/http/..."
		;;
	scheduler)
		printf '%s\n' "./internal/scheduler/service ./internal/scheduler/infra/grpc ./internal/scheduler/infra/kafka"
		;;
	users)
		printf '%s\n' "./internal/user/service ./internal/user/infra/http"
		;;
	*)
		printf 'Unknown service for coverage check: %s\n' "$1" >&2
		return 1
		;;
	esac
}

mkdir -p "$coverage_dir"
: > "$summary_file"

printf 'Service coverage threshold: %s%%\n' "$threshold" | tee -a "$summary_file"

status=0

for service in $services; do
	packages="$(packages_for_service "$service")"
	profile="$coverage_dir/$service.out"

	rm -f "$profile"

	printf '\n[%s]\n' "$service" | tee -a "$summary_file"
	go test -covermode=count -coverprofile="$profile" $packages

	coverage="$(go tool cover "-func=$profile" | awk '/^total:/ { sub(/%/, "", $3); print $3 }')"

	printf '%s coverage: %s%%\n' "$service" "$coverage" | tee -a "$summary_file"

	if awk -v coverage="$coverage" -v threshold="$threshold" 'BEGIN { exit !(coverage + 0 >= threshold + 0) }'; then
		printf 'OK: %s >= %s%%\n' "$service" "$threshold" | tee -a "$summary_file"
	else
		printf 'FAIL: %s coverage %s%% is below required %s%%\n' "$service" "$coverage" "$threshold" | tee -a "$summary_file"
		status=1
	fi
done

exit "$status"
