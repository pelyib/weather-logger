#!/usr/bin/env bash

# Setup CouchDB 
# script parameters: admin credentials, .env config file
# .env should define 
# - 2 DBs: api_raw_response, metrics
# - 1 couchdb user (it is called "logger" ) that has write access to both
# - 1 couchdb user (it is called "ui" ) that has read access to metrics only
# .env file:
# COUCHDB_HOST=
# COUCHDB_DB_NAME_API_RAW_RESPONSE=
# COUCHDB_DB_NAME_METRICS=
# COUCHDB_USER_NAME_LOGGER=
# COUCHDB_USER_PASSWORD_LOGGER=
# COUCHDB_USER_NAME_UI=
# COUCHDB_USER_PASSWORD_UI=

function ensure_user_created () {
    local user_name=$1
    local user_password=$2
    local roles="$3"

    echo "ensure '${user_name}' user exists with roles ${roles}"

    is_user_already_created=$(curl \
        --request GET \
        -H "${AUTH_HEADER}" \
        -o /dev/null -s -w "%{http_code}" \
        "$COUCHDB_HOST/_users/org.couchdb.user:${user_name}")

    if [[ ${is_user_already_created} -eq 404 ]]; then
        creation_result=$(curl \
            --request PUT \
            -H "${AUTH_HEADER}" \
            -H "Referer: ${COUCHDB_HOST}" \
            -H "Accept: application/json" \
            -H "Content-Type: application/json" \
            -o /dev/null -s -w "%{http_code}" \
            -d "{\"name\": \"${user_name}\", \"password\": \"${user_password}\", \"roles\": ${roles}, \"type\": \"user\"}" \
            "$COUCHDB_HOST/_users/org.couchdb.user:${user_name}")

        if [[ ${creation_result} -ne 201 ]]; then
            echo " └─'${user_name}' creation failed"
            exit 1
        fi

        echo " └─'${user_name}' created successfully"
        return
    fi

    echo " └─'${user_name}' already exists"
}

function ensure_db_created() {
    local db_name=$1

    echo "ensure '${db_name}' is created"

    is_db_already_created=$(curl \
        --request HEAD \
        -H "${AUTH_HEADER}" \
        -o /dev/null -s -w "%{http_code}" \
        -m 1 \
        "$COUCHDB_HOST/${db_name}")

    if [[ ${is_db_already_created} -eq 404 ]]; then
        creation_result=$(curl \
            --request PUT \
            -H "${AUTH_HEADER}" \
            -o /dev/null -s -w "%{http_code}" \
            -m 1 \
            "${COUCHDB_HOST}/${db_name}")

        if [[ ${creation_result} -ne 201 ]]; then
            echo " └─'${db_name}' creation failed"
            exit 1
        fi

        echo " └─'${db_name}' created successfully"
        return
    fi

    echo " └─'${db_name}' already exists"
}

function ensure_db_roles {
    db_name=$1
    role=$2

    echo "ensure '${role}' role exists in ${db_name}"

    # fetch current _security
    # push role to the $.members.roles array
    # put 
    response=$(curl \
        --request GET \
        -H "${AUTH_HEADER}" \
        -s -w "\n%{http_code}" \
        "${COUCHDB_HOST}/${db_name}/_security" \
    )

    body=$(echo "$response" | sed '$d')
    status_code=$(echo "$response" | tail -n1)

    updated_json=$(echo "$body" | jq -c --arg role "$role" '.members.roles |= if index($role) == null then . + [$role] else . end')

    if [[ "$body" == "$updated_json" ]]; then
        echo " └─${role} already exists in ${db_name}"
        return
    fi

    put_result=$(curl \
        --request PUT \
        -H "${AUTH_HEADER}" \
        -H "Accept: application/json" \
        -H "Content-Type: application/json" \
        -d "${updated_json}" \
        -o /dev/null -s -w "%{http_code}" \
        "${COUCHDB_HOST}/${db_name}/_security" \
    )

        if [[ ${put_result} -ne 200 ]]; then
            echo " └─adding '${role}' to ${db_name} failed"
            exit 1
        fi

        echo " └─adding '${role}' to ${db_name} succeeded"
        return
}

function ensure_valdocfunc_created() {
    db_name=$1

    echo "ensure validate_doc_update in '${db_name}' exists"

    creation_result=$(curl \
        --request PUT \
        -H "${AUTH_HEADER}" \
        -d '{"validate_doc_update": "function(newDoc, oldDoc, userCtx) { if (userCtx.roles.includes(\"wl.metrics.rw\")) { return; } throw({forbidden: \"not able now!\" });}"}' \
        -o /dev/null -s -w "%{http_code}" \
        "${COUCHDB_HOST}/${db_name}/_design/only-rw-can-write"
    )

    if [[ ${creation_result} -ne 201 ]]; then
        echo " └─validate_doc_update in '${db_name}' creation failed"
        exit 1
    fi

    echo " └─validate_doc_update in '${db_name}' created successfully"
    return
}

# check if script received 3 parameters
if [[ $# -ne 3 ]]; then
    echo "Usage: $0 <admin_user> <admin_password> <.env>"
    exit 1
fi

ADMIN_USER=$1
ADMIN_PASSWORD=$2

if [[ -z "$ADMIN_USER" || -z "$ADMIN_PASSWORD" ]]; then
    echo "admin credentials not provided"
    exit 1
fi

AUTH_HEADER="Authorization: Basic $(echo -n "$ADMIN_USER:$ADMIN_PASSWORD" | base64)"

. $3

if [[ 
        -z "$COUCHDB_HOST" ||
        -z "$COUCHDB_DB_NAME_API_RAW_RESPONSES" ||
        -z "$COUCHDB_DB_NAME_METRICS" ||
        -z "$COUCHDB_USER_NAME_LOGGER" || 
        -z "$COUCHDB_USER_PASSWORD_LOGGER" ||
        -z "$COUCHDB_USER_NAME_UI" || 
        -z "$COUCHDB_USER_PASSWORD_UI"
    ]]; then
    echo ".env file is invalid, one or more parameters are missing"
    exit 1
fi

echo ".env file is valid"

test_connection=$(curl \
    --request GET \
    -H "${AUTH_HEADER}" \
    -o /dev/null -s -w "%{http_code}" \
    "$COUCHDB_HOST")

if [[ ${test_connection} -ne 200 ]]; then
    echo "failed to connect to couchdb"
    exit 1
fi

echo "admin credentials are valid"

ensure_user_created $COUCHDB_USER_NAME_LOGGER $COUCHDB_USER_PASSWORD_LOGGER "[\"wl.metrics.rw\", \"wl.raw_responses.rw\"]"
ensure_user_created $COUCHDB_USER_NAME_UI $COUCHDB_USER_PASSWORD_UI "[\"wl.metrics.ro\"]"

echo ""
echo "users are ready"
echo ""

ensure_db_created "${COUCHDB_DB_NAME_API_RAW_RESPONSES}"
ensure_db_created "${COUCHDB_DB_NAME_METRICS}"

echo ""
echo "dbs are ready"
echo ""

ensure_db_roles "${COUCHDB_DB_NAME_API_RAW_RESPONSES}" "wl.raw_responses.rw"
ensure_db_roles "${COUCHDB_DB_NAME_METRICS}" "wl.metrics.rw"
ensure_db_roles "${COUCHDB_DB_NAME_METRICS}" "wl.metrics.ro"

ensure_valdocfunc_created "metrics"
