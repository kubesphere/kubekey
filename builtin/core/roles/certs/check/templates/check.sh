#!/bin/bash

# example
# ./check_registry.sh registry.example.com [options]
# return: 0 (success) or 1 (fail)

REGISTRY_HOST=""
REGISTRY_PORT=443
CA_FILE=""
CERT_FILE=""
KEY_FILE=""
INSECURE=false

parse_arguments() {
    REGISTRY_HOST=$1
    shift
    while [[ $# -gt 0 ]]; do
        case $1 in
            -p|--port)
                REGISTRY_PORT="$2"
                shift 2
                ;;
            --ca-file)
                CA_FILE="$2"
                shift 2
                ;;
            --cert-file)
                CERT_FILE="$2"
                shift 2
                ;;
            --key-file)
                KEY_FILE="$2"
                shift 2
                ;;
            -k|--insecure)
                INSECURE=true
                shift
                ;;
            -h|--help)
                exit 0
                ;;
            *)
                exit 1
                ;;
        esac
    done
    if [ -z "$REGISTRY_HOST" ]; then
        exit 1
    fi
}

# verify input ca cert key
validate_cert_files() {
    if [ -n "$CA_FILE" ] && [ ! -f "$CA_FILE" ]; then
        return 1
    fi
    if [ -n "$CERT_FILE" ] && [ ! -f "$CERT_FILE" ]; then
        return 1
    fi
    if [ -n "$KEY_FILE" ] && [ ! -f "$KEY_FILE" ]; then
        return 1
    fi
    return 0
}

# verify port accessible
check_port_connectivity() {
    local host=$1
    local port=$2

    if command -v nc >/dev/null 2>&1; then
        if nc -z -w 5 "$host" "$port" >/dev/null 2>&1; then
            return 0
        else
            return 1
        fi
    elif command -v timeout >/dev/null 2>&1; then
        if timeout 5 bash -c "echo >/dev/tcp/$host/$port" 2>/dev/null; then
            return 0
        else
            return 1
        fi
    else
        return 0
    fi
}

# check Registry API usable
check_registry_api() {
    local host=$1
    local port=$2
    local ca_file=$3
    local cert_file=$4
    local key_file=$5
    local insecure=$6

    if ! command -v curl >/dev/null 2>&1; then
        return 1
    fi

    local curl_cmd="curl -s --max-time 10"

    if [ -n "$ca_file" ]; then
        curl_cmd="$curl_cmd --cacert $ca_file"
    elif [ "$insecure" = true ]; then
        curl_cmd="$curl_cmd --insecure"
    fi

    if [ -n "$cert_file" ] && [ -n "$key_file" ]; then
        curl_cmd="$curl_cmd --cert $cert_file --key $key_file"
    fi

    local protocol="https"
    if [ "$port" = "80" ] || [ "$insecure" = true ]; then
        protocol="http"
    fi

    if eval "$curl_cmd $protocol://$host:$port/v2/" >/dev/null 2>&1; then
        return 0
    fi

    if [ "$protocol" = "https" ] && [ "$insecure" = true ]; then
        if eval "$curl_cmd http://$host:$port/v2/" >/dev/null 2>&1; then
            return 0
        fi
    fi

    local response
    response=$(eval "$curl_cmd -I $protocol://$host:$port/ 2>/dev/null | head -n 1 | cut -d' ' -f2")

    if [ -n "$response" ]; then
        return 0
    fi

    return 1
}

main() {
    parse_arguments "$@" || exit 1

    validate_cert_files || exit 1

    check_port_connectivity "$REGISTRY_HOST" "$REGISTRY_PORT" || exit 1

    check_registry_api "$REGISTRY_HOST" "$REGISTRY_PORT" "$CA_FILE" "$CERT_FILE" "$KEY_FILE" "$INSECURE" || exit 1

    exit 0
}

main "$@" 2>/dev/null